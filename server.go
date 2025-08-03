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
	r Router

	bufPool         sync.Pool
	writerPool      sync.Pool
	responseBufPool sync.Pool

	headerPool         sync.Pool
	responseHeaderPool sync.Pool
	requestPool        sync.Pool

	addr string

	maxHeaderSize int
}

// NewServer
func NewServer(opts ...ServerOption) Server {
	var unsafe unsafeServer
	// set defaults
	unsafe.maxHeaderSize = 16
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
	return func(s *unsafeServer) {
		s.addr = addr
	}
}

// WithRouter
func WithRouter(r Router) ServerOption {
	return func(s *unsafeServer) {
		s.r = r
	}
}

// WithMaxHeaderSize
func WithMaxHeaderSize(headerSize int) ServerOption {
	return func(s *unsafeServer) {
		s.maxHeaderSize = headerSize
	}
}

// Serve
func (s *unsafeServer) Serve(ctx context.Context) error {
	addr, err := net.ResolveTCPAddr("tcp", s.addr)
	if err != nil {
		return fmt.Errorf("net.ResolveTCPAddr: %w", err)
	}

	lis, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("net.Listen: %s %v", s.addr, err)
	}

	return s.listen(ctx, lis)
}

func (s *unsafeServer) listen(ctx context.Context, lis *net.TCPListener) error {
	var wg sync.WaitGroup

	numWorkers := runtime.NumCPU()
	fmt.Println(numWorkers)
	ch := make(chan net.Conn, numWorkers)

	go func() {
		<-ctx.Done()
		close(ch)
		lis.Close()
	}()

	for range numWorkers {
		wg.Add(1)
		go s.handleConnWorker(ctx, ch, &wg)
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

		if err := conn.SetReadDeadline(time.Now().Add(time.Second * 15)); err != nil {
			conn.Close()
			continue
		}

		ch <- conn

		// go func() {
		// 	defer wg.Done()
		// 	if err := s.handleConn(ctx, conn, &wg); err != nil {
		// 		switch err {
		// 		case ErrUnsupportedHttpVersion:
		// 			writeError(conn, StatusHTTPVersionNotSupported)
		// 		case ErrRouteNotFound:
		// 			writeError(conn, StatusNotFound)
		// 		default:
		// 			writeError(conn, StatusInternalServer)
		// 		}
		// 	}
		// }()
	}

SHUTDOWN:
	wg.Wait()

	return nil
}

func (s *unsafeServer) handleConnWorker(ctx context.Context, ch chan net.Conn, wg *sync.WaitGroup) error {
	defer wg.Done()
	for conn := range ch {
		if err := s.handleConn(ctx, conn); err != nil {
			writeError(conn, StatusInternalServer)
		}
	}
	return nil
}

func (s *unsafeServer) handleConn(ctx context.Context, conn net.Conn) error {
	defer conn.Close()

	buf := s.bufPool.Get().(*bufio.Reader)
	buf.Reset(conn)

	defer func() {
		buf.Reset(nil)
		s.bufPool.Put(buf)
	}()

	for {
		if err := conn.SetReadDeadline(time.Now().Add(time.Second * 15)); err != nil {
			return err
		}

		req, err := s.parseRequestFromBuf(buf)
		if err != nil {
			if req != nil {
				s.putRequestPool(req)
			}
			if err == io.EOF {
				return nil
			}
			return err
		}
		req.ctx = ctx

		w := s.newWriter(conn, req)

		if handle, ok := s.r.matchRoute(req); ok {
			handle(w, req)

			if err := w.writeResponse(); err != nil {
				writeError(w.conn, StatusInternalServer)
			}
		}

		s.putWriterPool(w)
		s.putRequestPool(req)
	}
}

func (s *unsafeServer) newWriter(conn net.Conn, request *Request) *writer {
	w := s.writerPool.Get().(*writer)

	buf := s.responseBufPool.Get().(*bytes.Buffer)
	buf.Reset()

	headers := s.responseHeaderPool.Get().(map[string]string)
	w.headers = headers

	w.conn = conn
	w.req = request
	w.statusCode = 0
	w.msg = ""
	w.body = buf

	return w
}

func (s *unsafeServer) putWriterPool(w *writer) {

	if w.headers != nil {
		for key := range w.headers {
			delete(w.headers, key)
		}
		s.responseHeaderPool.Put(w.headers)
	}

	if w.body != nil {
		w.body.Reset()
		s.responseBufPool.Put(w.body)
	}

	w.conn = nil
	w.req = nil
	w.body = nil

	s.writerPool.Put(w)
}

func (s *unsafeServer) putRequestPool(r *Request) {
	if r == nil {
		return
	}

	if r.Headers != nil {
		for k := range r.Headers {
			delete(r.Headers, k)
		}
		s.headerPool.Put(r.Headers)
	}

	r.ctx = nil
	if r.Body != nil {
		r.Body.Reset()
		s.responseBufPool.Put(r.Body)
	}
	r.Body = nil

	r.Method = nil
	r.Params = nil
	r.Path = nil
	r.Version = nil

	s.requestPool.Put(r)
}

func writeError(conn net.Conn, code StatusCode) {
	msg := code.String()
	fmt.Fprintf(conn, "HTTP/1.1 %d %s\r\nContent-Length: %d\r\n\r\n%s", code, msg, len(msg), msg)
}
