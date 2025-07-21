package unsafehttp

import (
	"fmt"
	"net"
)

// ResponseWriter
type ResponseWriter interface {
	SetStatus(StatusCode)
	SetHeader(string, string)
	Write([]byte) (int, error)
}

type writer struct {
	conn       net.Conn
	req        *Request
	statusCode StatusCode
	msg        string

	headers Headers
	body    []byte
}

// interface compliance
var _ ResponseWriter = (*writer)(nil)

func newResponseWriter(conn net.Conn, request *Request) *writer {
	return &writer{
		conn:    conn,
		req:     request,
		body:    []byte{},
		headers: make(Headers),
	}
}

// SetStatus
func (w *writer) SetStatus(code StatusCode) {
	w.statusCode = code
	w.msg = code.String()
}

// SetHeader
func (w *writer) SetHeader(key, value string) {
	w.headers[key] = value
}

// Write
func (w *writer) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return len(b), nil
}

func (w *writer) writeResponse() error {

	// Status-Line = HTTP-Version SP Status-Code SP Reason-Phrase CRLF
	statusLine := fmt.Sprintf("%s %d %s\r\n", w.req.Version, w.statusCode, w.msg)
	_, err := w.conn.Write([]byte(statusLine))
	if err != nil {
		return fmt.Errorf("unable to write status line conn.Write: %w", err)
	}

	// set default header values if not already set
	if _, ok := w.headers["Content-Type"]; !ok {
		w.headers["Content-Type"] = "text/plain; charset=utf-8"
	}

	if _, ok := w.headers["Content-Length"]; !ok {
		w.headers["Content-Length"] = fmt.Sprintf("%d", len(w.body))
	}

	for key, value := range w.headers {
		headerLine := fmt.Sprintf("%s:%s\r\n", key, value)
		_, err = w.conn.Write([]byte(headerLine))
		if err != nil {
			return fmt.Errorf("unable to write header line conn.Write: %w", err)
		}
	}

	// write line separate headers and body
	_, err = w.conn.Write([]byte("\r\n"))
	if err != nil {
		return fmt.Errorf("failed to write header-body separator: %w", err)
	}

	_, err = w.conn.Write(w.body)
	if err != nil {
		return fmt.Errorf("unable to write response body: %w", err)
	}
	defer w.conn.Close()

	return nil
}
