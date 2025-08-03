package unsafehttp

import (
	"bytes"
	"fmt"
	"net"
	"strconv"
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

	headers map[string]string
	body    *bytes.Buffer
}

// interface compliance
var _ ResponseWriter = (*writer)(nil)

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
	return w.body.Write(b)
}

func (w *writer) writeResponse() error {
	var buf bytes.Buffer

	fmt.Fprintf(&buf, "%s %d %s\r\n", w.req.Version, w.statusCode, w.msg)
	if _, ok := w.headers["Content-Type"]; ok {
		w.headers["Content-Type"] = "text/plain; charset=utf-8"
	}
	w.headers["Content-Length"] = strconv.Itoa(w.body.Len())

	for key, value := range w.headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", key, value)
	}
	buf.WriteString("\r\n")

	_, err := buf.Write(w.body.Bytes())
	if err != nil {
		return fmt.Errorf("buf.Write: %w", err)
	}

	_, err = w.conn.Write(buf.Bytes())
	if err != nil {
		return fmt.Errorf("conn.Write: %w", err)
	}

	return nil
}
