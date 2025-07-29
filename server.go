package unsafehttp

import (
	"bufio"
	"context"
	"fmt"
	"net"
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
	mux  *Mux
	addr string
}

// NewServer
func NewServer(opts ...ServerOption) Server {
	var unsafe unsafeServer

	for _, opt := range opts {
		opt(&unsafe)
	}

	return &unsafe
}

// WithAddr
func WithAddr(addr string) ServerOption {
	return func(s *unsafeServer) {
		s.addr = addr
	}
}

// WithMux
func WithMux(m *Mux) ServerOption {
	return func(s *unsafeServer) {
		s.mux = m
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

	s.listen(ctx, lis)
	return nil
}

func (s *unsafeServer) listen(ctx context.Context, lis *net.TCPListener) {
	defer lis.Close()

	var wg sync.WaitGroup

	// add for server
	wg.Add(1)

OUTER:
	for {
		select {
		case <-ctx.Done():
			break OUTER
		default:
			// fallthrough
		}

		if err := lis.SetDeadline(time.Now().Add(time.Millisecond * 250)); err != nil {
			// unable to set listener deadline
			// sleep and retry
			time.Sleep(time.Second)
			continue
		}

		conn, err := lis.Accept()
		if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
			// listener timeout continue loop
			continue
		} else if err != nil {
			writeError(conn, StatusInternalServer)
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := s.handleConn(ctx, conn); err != nil {
				switch err {
				case ErrUnsupportedHttpVersion:
					writeError(conn, StatusHTTPVersionNotSupported)
				case ErrRouteNotFound:
					writeError(conn, StatusNotFound)
				default:
					writeError(conn, StatusInternalServer)
				}
			}
		}()
	}

	// server shutdown
	wg.Done()

	// wait for all existing
	wg.Wait()
}

func (s *unsafeServer) handleConn(ctx context.Context, conn net.Conn) error {

	r := bufio.NewReader(conn)

	req, err := parseRequestFromBuf(r)
	if err != nil {
		return err
	}
	req.ctx = ctx

	rw := newResponseWriter(conn, req)

	if handle, ok := s.mux.matchRoute(req); ok {
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			handle.ServeHTTP(rw, req)
			if err := rw.writeResponse(); err != nil {
				writeError(rw.conn, StatusInternalServer)
			}
		}()
		wg.Wait()
		return nil
	}

	return ErrRouteNotFound
}

func writeError(conn net.Conn, code StatusCode) {
	errMsg := fmt.Sprintf("%s %d %s\r\n\r\n\r\n", "HTTP/1.1", code, code.String())
	conn.Write([]byte(errMsg))
}
