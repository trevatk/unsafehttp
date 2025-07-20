package unsafehttp

import (
	"fmt"
	"net"
)

// ResponseWriter
type ResponseWriter interface {
	SetStatus(StatusCode)
}

type writer struct {
	conn       net.Conn
	req        *Request
	statusCode StatusCode
	msg        string

	Headers Headers
	Body    []byte
}

// interface compliance
var _ ResponseWriter = (*writer)(nil)

func newResponseWriter(conn net.Conn, request *Request) *writer {
	return &writer{
		conn:    conn,
		req:     request,
		Body:    []byte{},
		Headers: make(Headers),
	}
}

// SetStatus
func (w *writer) SetStatus(code StatusCode) {
	w.statusCode = code
	w.msg = code.String()
}

func (w *writer) writeResponse() error {

	// Status-Line = HTTP-Version SP Status-Code SP Reason-Phrase CRLF
	statusLine := fmt.Sprintf("%s %d %s\r\n", w.req.Version, w.statusCode, w.msg)
	_, err := w.conn.Write([]byte(statusLine))
	if err != nil {
		return fmt.Errorf("unable to write status line conn.Write: %w", err)
	}

	// set default header values if not already set
	if _, ok := w.Headers["Content-Type"]; !ok {
		w.Headers["Content-Type"] = "text/plain; charset=utf-8"
	}

	if _, ok := w.Headers["Content-Length"]; !ok {
		w.Headers["Content-Length"] = fmt.Sprintf("%d", len(w.Body))
	}

	for key, value := range w.Headers {
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

	_, err = w.conn.Write(w.Body)
	if err != nil {
		return fmt.Errorf("unable to write response body: %w", err)
	}
	defer w.conn.Close()

	return nil
}
