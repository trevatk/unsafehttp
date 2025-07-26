package unsafehttp

import "strings"

// Method
type Method string

// StatusCode
type StatusCode int

// String
func (sc StatusCode) String() string {
	switch sc {
	case StatusContinue:
		return "continue"
	case StatusSwitchProtocol:
		return "upgrade"
	case StatusCreated:
		return "created"
	case StatusAccepted:
		return "accepted"
	case StatusNonAuthoritative:
		return "nonauthoritative"
	case StatusNoContent:
		return "no content"
	case StatusResetContent:
		return "reset content"
	case StatusPartialContent:
		return "partial content"
	case StatusMultipleChoices:
		return "multiple choices"
	case StatusMovedPermanently:
		return "moved permanently"
	case StatusFound:
		return "found"
	case StatusSeeOther:
		return "see other"
	case StatusNotModified:
		return "not modified"
	case StatusUseProxy:
		return "use proxy"
	case StatusUnused:
		return "unused"
	case StatusTemporaryRedirect:
		return "redirect"
	case StatusBadRequest:
		return "bad request"
	case StatusUnauthorized:
		return "unauthorized"
	case StatusPaymentRequired:
		return "payment required"
	case StatusForbidden:
		return "403"
	case StatusNotFound:
		return "not found"
	case StatusMethodNotAllowed:
		return "method not allowed"
	case StatusNotApplicable:
		return "not applicable"
	case StatusProxyAuthenticationRequired:
		return "proxy authenitcation required"
	case StatusRequestTimeout:
		return "request timeout"
	case StatusConflict:
		return "conflict"
	case StatusGone:
		return "gone"
	case StatusLengthRequired:
		return "length required"
	case StatusPreconditionFailed:
		return "precondition failed"
	case StatusRequestEntityTooLarge:
		return "request entity too large"
	case StatusUnsupportedMediaType:
		return "unsupported media type"
	case StatusRequestedRangeNotSatisfiable:
		return "requested range not satisfiable"
	case StatusExpectationFailed:
		return "expectation failed"
	case StatusTeapot:
		return "teapot"
	case StatusInternalServer:
		return "internal server"
	case StatusNotImplemented:
		return "not implemented"
	case StatusBadGateway:
		return "bad gateway"
	case StatusServiceUnavailable:
		return "service unavailable"
	case StatusGatewayTimeout:
		return "gateway timeout"
	case StatusHTTPVersionNotSupported:
		return "http version not supported"
	default:
		return "ok"
	}
}

func methodfromString(s string) (Method, bool) {
	switch strings.ToLower(s) {
	case "get":
		return MethodGet, true
	case "post":
		return MethodPost, true
	case "put":
		return MethodPut, true
	case "patch":
		return MethodPatch, true
	case "delete":
		return MethodDelete, true
	case "head":
		return MethodHead, true
	case "options":
		return MethodOptions, true
	case "connect":
		return MethodConnect, true
	default:
		return "", false
	}
}

const (
	// GET
	MethodGet Method = "GET"
	// POST
	MethodPost Method = "POST"
	// Put
	MethodPut Method = "PUT"
	// Patch
	MethodPatch Method = "PATCH"
	// Delete
	MethodDelete Method = "DELETE"
	// Head
	MethodHead Method = "HEAD"
	// Options
	MethodOptions Method = "OPTIONS"
	// Connect
	MethodConnect Method = "CONNECT"
	// Trace
	MethodTrace Method = "TRACE"

	// 1xx

	StatusContinue       StatusCode = 100
	StatusSwitchProtocol StatusCode = 101

	// 2xx

	StatusOK               StatusCode = 200
	StatusCreated          StatusCode = 201
	StatusAccepted         StatusCode = 202
	StatusNonAuthoritative StatusCode = 203
	StatusNoContent        StatusCode = 204
	StatusResetContent     StatusCode = 205
	StatusPartialContent   StatusCode = 206

	// 3xx

	StatusMultipleChoices   StatusCode = 300
	StatusMovedPermanently  StatusCode = 301
	StatusFound             StatusCode = 302
	StatusSeeOther          StatusCode = 303
	StatusNotModified       StatusCode = 304
	StatusUseProxy          StatusCode = 305
	StatusUnused            StatusCode = 306
	StatusTemporaryRedirect StatusCode = 307

	// 4xx

	StatusBadRequest                   StatusCode = 400
	StatusUnauthorized                 StatusCode = 401
	StatusPaymentRequired              StatusCode = 402
	StatusForbidden                    StatusCode = 403
	StatusNotFound                     StatusCode = 404
	StatusMethodNotAllowed             StatusCode = 405
	StatusNotApplicable                StatusCode = 406
	StatusProxyAuthenticationRequired  StatusCode = 407
	StatusRequestTimeout               StatusCode = 408
	StatusConflict                     StatusCode = 409
	StatusGone                         StatusCode = 410
	StatusLengthRequired               StatusCode = 411
	StatusPreconditionFailed           StatusCode = 412
	StatusRequestEntityTooLarge        StatusCode = 413
	StatusRequestUriToLong             StatusCode = 414
	StatusUnsupportedMediaType         StatusCode = 415
	StatusRequestedRangeNotSatisfiable StatusCode = 416
	StatusExpectationFailed            StatusCode = 417
	StatusTeapot                       StatusCode = 418

	// 5xx

	StatusInternalServer          StatusCode = 500
	StatusNotImplemented          StatusCode = 501
	StatusBadGateway              StatusCode = 502
	StatusServiceUnavailable      StatusCode = 503
	StatusGatewayTimeout          StatusCode = 504
	StatusHTTPVersionNotSupported StatusCode = 505
)
