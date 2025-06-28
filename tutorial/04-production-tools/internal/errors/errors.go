package errors

import (
	"fmt"
	"runtime"

	"github.com/pkg/errors"
)

// ErrorCode represents different types of errors
type ErrorCode string

const (
	ErrCodeInvalidInput    ErrorCode = "INVALID_INPUT"
	ErrCodeFileNotFound    ErrorCode = "FILE_NOT_FOUND"
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"
	ErrCodeProcessingError ErrorCode = "PROCESSING_ERROR"
	ErrCodeNetworkError    ErrorCode = "NETWORK_ERROR"
	ErrCodeConfigError     ErrorCode = "CONFIG_ERROR"
	ErrCodeInternalError   ErrorCode = "INTERNAL_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
	Cause   error                  `json:"-"`
	Stack   string                 `json:"stack,omitempty"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// New creates a new AppError
func New(code ErrorCode, message string, details ...map[string]interface{}) *AppError {
	err := &AppError{
		Code:    code,
		Message: message,
		Stack:   getStack(),
	}
	
	if len(details) > 0 {
		err.Details = details[0]
	}
	
	return err
}

// Wrap wraps an existing error with additional context
func Wrap(err error, code ErrorCode, message string, details ...map[string]interface{}) *AppError {
	appErr := &AppError{
		Code:    code,
		Message: message,
		Cause:   err,
		Stack:   getStack(),
	}
	
	if len(details) > 0 {
		appErr.Details = details[0]
	}
	
	return appErr
}

// IsErrorCode checks if an error has a specific error code
func IsErrorCode(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func getStack() string {
	const depth = 32
	var pcs [depth]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])
	
	var stack string
	for {
		frame, more := frames.Next()
		stack += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
		if !more {
			break
		}
	}
	return stack
}
