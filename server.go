package unsafehttp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"runtime"
	"sync"
	"time"
)

// ServerOption
type ServerOption func(*unsafeServer)

// Server
type Server interface {
	Serve(context.Context) error
}

type unsafeServer struct {
	r *router

	bufPool         sync.Pool
	writerPool      sync.Pool
	responseBufPool sync.Pool

	headerPool         sync.Pool
	responseHeaderPool sync.Pool
	requestPool        sync.Pool

	addr string

	maxHeaderSize int
	maxBodySize   int64

	// concurrency
	numWorkers int

	// connection deadlines
	connReadTimeout  time.Duration
	connWriteTimeout time.Duration
	connTimeout      time.Duration
}

// NewServer
func NewServer(opts ...ServerOption) Server {
	var unsafe unsafeServer
	// set defaults
	unsafe.maxHeaderSize = 16
	unsafe.numWorkers = runtime.NumCPU()
	unsafe.maxBodySize = 10000000

	unsafe.connReadTimeout = time.Second * 15
	unsafe.connWriteTimeout = time.Second * 15
	unsafe.connTimeout = time.Second * 60

	for _, opt := range opts {
		opt(&unsafe)
	}

	unsafe.bufPool = sync.Pool{
		New: func() any {
			return new(bufio.Reader)
		},
	}

	unsafe.responseBufPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}

	unsafe.writerPool = sync.Pool{
		New: func() any {
			return new(writer)
		},
	}

	unsafe.headerPool = sync.Pool{
		New: func() any {
			return make(map[string]string, unsafe.maxHeaderSize)
		},
	}

	unsafe.responseHeaderPool = sync.Pool{
		New: func() any {
			return make(map[string]string, unsafe.maxHeaderSize)
		},
	}

	unsafe.requestPool = sync.Pool{
		New: func() any {
			return new(Request)
		},
	}

	return &unsafe
}

// WithAddr
func WithAddr(addr string) ServerOption {
	return func(us *unsafeServer) {
		us.addr = addr
	}
}

// WithRouter
func WithRouter(r Router) ServerOption {
	return func(us *unsafeServer) {
		us.r = r.(*router)
	}
}

// WithMaxHeaderSize
func WithMaxHeaderSize(headerSize int) ServerOption {
	return func(us *unsafeServer) {
		us.maxHeaderSize = headerSize
	}
}

func WithMaxBodySize(bodySize int64) ServerOption {
	return func(us *unsafeServer) {
		us.maxBodySize = bodySize
	}
}

// WithConcurrency
func WithConcurrency(numWorkers int) ServerOption {
	return func(us *unsafeServer) {
		us.numWorkers = numWorkers
	}
}

// WithConnTimeout
func WithConnTimeout(timeout time.Duration) ServerOption {
	return func(us *unsafeServer) {
		us.connTimeout = timeout
	}
}

// WithConnReadTimeout
func WithConnReadTimeout(timeout time.Duration) ServerOption {
	return func(us *unsafeServer) {
		us.connReadTimeout = timeout
	}
}

// WithConnWriteTimeout
func WithConnWriteTimeout(timeout time.Duration) ServerOption {
	return func(us *unsafeServer) {
		us.connWriteTimeout = timeout
	}
}

// Serve
func (us *unsafeServer) Serve(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr("tcp", us.addr)
	if err != nil {
		return fmt.Errorf("net.ResolveTCPAddr: %w", err)
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("net.Listen: %s %v", us.addr, err)
	}

	return us.listen(ctx, lis)
}

func (us *unsafeServer) listen(ctx context.Context, lis *net.TCPListener) error {
	var wg sync.WaitGroup

	ch := make(chan net.Conn, us.numWorkers)

	go func() {
		<-ctx.Done()
		close(ch)
		lis.Close()
	}()

	for range us.numWorkers {
		wg.Add(1)
		go us.handleConnWorker(ctx, ch, &wg)
	}

	for {
		conn, err := lis.Accept()

		if err != nil {
			if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
				continue
			}

			select {
			case <-ctx.Done():
				goto SHUTDOWN
			default:
				if conn != nil {
					conn.Close()
				}
				return fmt.Errorf("lis.Accept: %w", err)
			}
		}

		if err := lis.SetDeadline(time.Now().Add(time.Second * 15)); err != nil {
			conn.Close()
			continue
		}

		ch <- conn
	}

SHUTDOWN:
	wg.Wait()

	return nil
}

func (us *unsafeServer) handleConnWorker(ctx context.Context, ch chan net.Conn, wg *sync.WaitGroup) error {
	defer wg.Done()
	for conn := range ch {
		if err := us.handleConn(ctx, conn); err != nil {
			writeError(conn, StatusInternalServer)
		}
	}
	return nil
}

func (us *unsafeServer) handleConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	buf := us.bufPool.Get().(*bufio.Reader)
	buf.Reset(conn)

	defer func() {
		buf.Reset(nil)
		us.bufPool.Put(buf)
	}()

	for {

		conn.SetDeadline(time.Now().Add(us.connTimeout))
		if err := conn.SetReadDeadline(time.Now().Add(us.connReadTimeout)); err != nil {
			return err
		}

		conn.SetWriteDeadline(time.Now().Add(us.connWriteTimeout))

		req, err := us.parseRequestFromBuf(buf)
		if err != nil {
			if req != nil {
				us.putRequestPool(req)
			}
			if err == io.EOF {
				return nil
			}
			switch err {
			case ErrRequestBodyTooLarge:
				writeError(conn, StatusRequestEntityTooLarge)
			case ErrUnsupportedHttpVersion:
				writeError(conn, StatusHTTPVersionNotSupported)
			default:
				writeError(conn, StatusInternalServer)
			}
		}
		req.ctx = ctx

		w := us.newWriter(conn, req)

		if handle, ok := us.r.matchRoute(req); ok {
			handle.ServeHTTP(w, req)

			if err := w.writeResponse(); err != nil {
				writeError(w.conn, StatusInternalServer)
			}
		} else {
			writeError(conn, StatusNotFound)
		}

		us.putWriterPool(w)
		us.putRequestPool(req)
	}
}

func (us *unsafeServer) newWriter(conn net.Conn, request *Request) *writer {
	w := us.writerPool.Get().(*writer)

	buf := us.responseBufPool.Get().(*bytes.Buffer)
	buf.Reset()

	headers := us.responseHeaderPool.Get().(map[string]string)
	w.headers = headers

	w.conn = conn
	w.req = request
	w.statusCode = 0
	w.msg = ""
	w.body = buf

	return w
}

func (us *unsafeServer) putWriterPool(w *writer) {

	if w.headers != nil {
		for key := range w.headers {
			delete(w.headers, key)
		}
		us.responseHeaderPool.Put(w.headers)
	}

	if w.body != nil {
		w.body.Reset()
		us.responseBufPool.Put(w.body)
	}

	w.conn = nil
	w.req = nil
	w.body = nil

	us.writerPool.Put(w)
}

func (us *unsafeServer) putRequestPool(r *Request) {
	if r == nil {
		return
	}

	if r.Headers != nil {
		for k := range r.Headers {
			delete(r.Headers, k)
		}
		us.headerPool.Put(r.Headers)
	}

	r.ctx = nil
	if r.Body != nil {
		r.Body.Reset()
		us.responseBufPool.Put(r.Body)
	}
	r.Body = nil

	r.Method = nil
	r.Params = nil
	r.Path = nil
	r.Version = nil

	us.requestPool.Put(r)
}

func writeError(conn net.Conn, code StatusCode) {
	msg := code.String()
	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: %d\r\n\r\n%s", code, msg, len(msg), msg)
}
