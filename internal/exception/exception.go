package exception

import "fmt"

type Code int

const (
	CodeSuccess Code = 0
	CodeUnknown Code = 1000

	CodeBadRequest     Code = 4000
	CodeUnauthorized   Code = 4001
	CodeForbidden      Code = 4003
	CodeNotFound       Code = 4004
	CodeConflict       Code = 4009
	CodeValidation     Code = 4002
	CodeInvalidToken   Code = 4010
	CodeExpiredToken   Code = 4011
	CodeInvalidPassword Code = 4020

	CodeDatabase Code = 5001
	CodeInternal Code = 5000
)

type Exception struct {
	Code    Code   `json:"code"`
	Message string `json:"message"`
	Cause   error  `json:"-"`
}

func (e *Exception) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

func (e *Exception) Unwrap() error { return e.Cause }

func New(code Code, msg string) *Exception {
	return &Exception{Code: code, Message: msg}
}

func Newf(code Code, format string, args ...interface{}) *Exception {
	return &Exception{Code: code, Message: fmt.Sprintf(format, args...)}
}

func Wrap(code Code, msg string, cause error) *Exception {
	return &Exception{Code: code, Message: msg, Cause: cause}
}

func NewBadRequest(msg string) *Exception     { return New(CodeBadRequest, msg) }
func NewUnauthorized(msg string) *Exception   { return New(CodeUnauthorized, msg) }
func NewForbidden(msg string) *Exception      { return New(CodeForbidden, msg) }
func NewNotFound(msg string) *Exception       { return New(CodeNotFound, msg) }
func NewConflict(msg string) *Exception       { return New(CodeConflict, msg) }
func NewValidation(msg string) *Exception     { return New(CodeValidation, msg) }
func NewInvalidToken(msg string) *Exception   { return New(CodeInvalidToken, msg) }
func NewExpiredToken(msg string) *Exception   { return New(CodeExpiredToken, msg) }
func NewInvalidPassword(msg string) *Exception { return New(CodeInvalidPassword, msg) }
func NewDatabase(cause error) *Exception      { return Wrap(CodeDatabase, "database error", cause) }
func NewInternal(cause error) *Exception      { return Wrap(CodeInternal, "internal error", cause) }
