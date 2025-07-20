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

const (
	requestLineLen = 3
)

type readLine struct {
	method          Method
	target, version string
}

// Headers
type Headers map[string]string

func (h Headers) Set(key, value string) {
	h[key] = value
}

// Params
type Params map[string]string

// Request
type Request struct {
	Method  string
	Version string
	Path    string
	Headers Headers
	Params  Params
	Body    []byte

	ctx context.Context
}

// Context
func (r *Request) Context() context.Context {
	return r.ctx
}

func parseRequestFromBuf(reader *bufio.Reader) (*Request, error) {
	fl, err := reader.ReadBytes('\n')
	if err != nil {
		return nil, fmt.Errorf("buf.ReadBytes: %w", err)
	}

	rl, err := parseReadLine(fl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse read line: %w", err)
	}

	headers, err := parseHeaders(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to parse header: %w", err)
	}

	body, err := extractRequestBody(headers, reader)
	if err != nil {
		return nil, fmt.Errorf("failed to extract request body: %w", err)
	}

	return &Request{
		Method:  string(rl.method),
		Version: rl.version,
		Path:    rl.target,
		Headers: headers,
		Params:  make(Params),
		Body:    body,
	}, nil
}

func parseReadLine(line []byte) (readLine, error) {
	lineAsStr := string(line)

	// split line by SP
	s := strings.Split(lineAsStr, " ")
	if len(s) != requestLineLen {
		return readLine{}, fmt.Errorf("invalid number of fields in first line %d", len(s))
	}

	// parse string to method
	var m Method
	m = m.fromString(s[0])

	return readLine{
		method:  m,
		target:  strings.TrimSpace(s[1]),
		version: strings.TrimSpace(s[2]),
	}, nil
}

func parseHeaders(reader *bufio.Reader) (Headers, error) {
	headers := make(map[string]string, 0)

	for {
		line, err := reader.ReadBytes('\n')
		if err != nil {
			return Headers{}, fmt.Errorf("buf.ReadBytes: %w", err)
		}

		// line is clrf
		// end headers
		if bytes.Equal(line, []byte("\r\n")) {
			break
		}

		lineAsStr := string(line)
		s := strings.Split(lineAsStr, ":")

		key := strings.TrimSpace(s[0])
		value := strings.TrimSpace(s[1])
		headers[key] = value
	}

	return headers, nil
}

func extractRequestBody(headers Headers, reader *bufio.Reader) ([]byte, error) {
	if contentLengthStr, ok := headers["Content-Length"]; ok {
		contentLength, err := strconv.ParseInt(contentLengthStr, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("strconv.ParseInt: %w", err)
		}

		if contentLength < 0 {
			return nil, fmt.Errorf("negative content length value %d", contentLength)
		}

		if contentLength == 0 {
			// request is likely a GET or HEAD
			return nil, nil
		}

		body := make([]byte, contentLength)
		_, err = io.ReadFull(reader, body)
		if err != nil {
			return nil, fmt.Errorf("io.ReadFull: %w", err)
		}

		return body, nil
	}

	return nil, nil
}
