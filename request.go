package unsafehttp

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Request
type Request struct {
	Method  []byte
	Version []byte
	Path    []byte
	Headers map[string]string
	Params  map[string]string
	Body    *bytes.Buffer

	ctx context.Context
}

// Context
func (r *Request) Context() context.Context {
	return r.ctx
}

// WithContext
func (r *Request) WithContext(ctx context.Context) {
	r.ctx = ctx
}

func (us *unsafeServer) parseRequestFromBuf(buf *bufio.Reader) (*Request, error) {
	r := us.requestPool.Get().(*Request)

	statusLine, err := buf.ReadBytes('\n')
	if err != nil {
		if err == io.EOF {
			return nil, err
		}
		return nil, fmt.Errorf("unable to first line in buffer: %w", err)
	}

	// split status line
	ss := bytes.Split(statusLine, []byte(" "))
	r.Method = bytes.TrimSpace(ss[0])
	r.Path = bytes.TrimSpace(ss[1])
	r.Version = bytes.TrimSpace(ss[2])

	if !bytes.Equal(r.Version, HTTP1) && !bytes.Equal(r.Version, HTTP1_1) {
		return nil, ErrUnsupportedHttpVersion
	}

	// get pooled headers
	headers := us.headerPool.Get().(map[string]string)
	r.Headers = headers

	if err := parseHeaders(buf, headers); err != nil {
		// error found put request back to pool
		us.putRequestPool(r)
		return nil, err
	}

	body, err := us.extractRequestBody(headers, buf)
	if err != nil {
		us.putRequestPool(r)
		return nil, err
	}
	r.Body = body

	return r, nil
}

func parseHeaders(buf *bufio.Reader, headers map[string]string) error {
	for {
		line, err := buf.ReadBytes('\n')
		if err != nil {
			return fmt.Errorf("buf.ReadBytes: %w", err)
		}

		// line is "\r\n" which signifies the end of the headers
		if len(line) == 2 && bytes.Equal(line, []byte{'\r', '\n'}) {
			break
		}

		// Find the index of the first colon
		idx := bytes.IndexByte(line, ':')
		if idx == -1 {
			// Malformed header, could return an error or skip.
			// For robustness, we will skip this line.
			continue
		}

		// Extract and trim the key and value as byte slices
		key := bytes.TrimSpace(line[:idx])
		value := bytes.TrimSpace(line[idx+1:])

		// Convert to string once for map storage
		headers[string(key)] = string(value)
	}
	return nil
}

func (us *unsafeServer) extractRequestBody(headers map[string]string, buf *bufio.Reader) (*bytes.Buffer, error) {

	bb := us.responseBufPool.Get().(*bytes.Buffer)
	bb.Reset()

	if contentLengthStr, ok := headers["Content-Length"]; ok {
		contentLength, err := strconv.ParseInt(strings.TrimSpace(contentLengthStr), 10, 64)
		if err != nil {
			us.responseBufPool.Put(bb)
			return nil, fmt.Errorf("strconv.ParseInt: %w", err)
		}

		if contentLength <= 0 {
			return bb, nil
		}

		if contentLength > int64(us.maxBodySize) {
			return nil, ErrRequestBodyTooLarge
		}

		if _, err := io.CopyN(bb, buf, contentLength); err != nil {
			us.responseBufPool.Put(bb)

			return nil, fmt.Errorf("io.CopyN: %w", err)
		}
		return bb, nil
	}

	return bb, nil
}
