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
	case StatusBadRequest:
		return "Bad Request"
	case StatusInternalServer:
		return "Internal Server"
	default:
		return "OK"
	}
}

func (m Method) fromString(s string) Method {
	switch strings.ToLower(s) {
	case "get":
		return MethodGet
	case "post":
		return MethodPost
	case "put":
		return MethodPut
	case "patch":
		return MethodPatch
	case "delete":
		return MethodDelete
	case "head":
		return MethodHead
	case "options":
		return MethodOptions
	case "connect":
		return MethodConnect
	default:
		return MethodTrace
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
	StatusMethodNotallowed             StatusCode = 405
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
	StausTeapot                        StatusCode = 418

	// 5xx

	StatusInternalServer          StatusCode = 500
	StatusNotImplemented          StatusCode = 501
	StatusBadGateway              StatusCode = 502
	StatusServiceUnavailable      StatusCode = 503
	StatusGatewayTimeout          StatusCode = 504
	StatusHTTPVersionNotSupported StatusCode = 505
)
