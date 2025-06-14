package lsp

import (
	"encoding/json"
	"fmt"

	"github.com/sourcegraph/jsonrpc2"
)

// LSPErrorCode represents specific error codes for the LSP server
type LSPErrorCode int

const (
	// Standard JSON-RPC error codes
	ErrorCodeParseError     LSPErrorCode = -32700
	ErrorCodeInvalidRequest LSPErrorCode = -32600
	ErrorCodeMethodNotFound LSPErrorCode = -32601
	ErrorCodeInvalidParams  LSPErrorCode = -32602
	ErrorCodeInternalError  LSPErrorCode = -32603

	// LSP-specific error codes
	ErrorCodeServerNotInitialized LSPErrorCode = -32002
	ErrorCodeUnknownErrorCode     LSPErrorCode = -32001

	// Custom application error codes
	ErrorCodeDocumentNotFound    LSPErrorCode = -32100
	ErrorCodeInvalidDocument     LSPErrorCode = -32101
	ErrorCodeDocumentSyncFailed  LSPErrorCode = -32102
	ErrorCodeCompletionFailed    LSPErrorCode = -32103
	ErrorCodeHoverFailed         LSPErrorCode = -32104
	ErrorCodeDefinitionFailed    LSPErrorCode = -32105
	ErrorCodeReferencesFailed    LSPErrorCode = -32106
	ErrorCodeDocumentSymbolFailed LSPErrorCode = -32107
)

// String returns the string representation of the error code
func (code LSPErrorCode) String() string {
	switch code {
	case ErrorCodeParseError:
		return "ParseError"
	case ErrorCodeInvalidRequest:
		return "InvalidRequest"
	case ErrorCodeMethodNotFound:
		return "MethodNotFound"
	case ErrorCodeInvalidParams:
		return "InvalidParams"
	case ErrorCodeInternalError:
		return "InternalError"
	case ErrorCodeServerNotInitialized:
		return "ServerNotInitialized"
	case ErrorCodeUnknownErrorCode:
		return "UnknownErrorCode"
	case ErrorCodeDocumentNotFound:
		return "DocumentNotFound"
	case ErrorCodeInvalidDocument:
		return "InvalidDocument"
	case ErrorCodeDocumentSyncFailed:
		return "DocumentSyncFailed"
	case ErrorCodeCompletionFailed:
		return "CompletionFailed"
	case ErrorCodeHoverFailed:
		return "HoverFailed"
	case ErrorCodeDefinitionFailed:
		return "DefinitionFailed"
	case ErrorCodeReferencesFailed:
		return "ReferencesFailed"
	case ErrorCodeDocumentSymbolFailed:
		return "DocumentSymbolFailed"
	default:
		return "UnknownError"
	}
}

// LSPError represents a custom LSP error with additional context
type LSPError struct {
	Code    LSPErrorCode
	Message string
	Data    interface{}
	Cause   error
	Context map[string]interface{}
}

// Error implements the error interface
func (e *LSPError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s (%s): %s - caused by: %v", e.Code.String(), e.Message, e.formatContext(), e.Cause)
	}
	return fmt.Sprintf("%s (%s): %s", e.Code.String(), e.Message, e.formatContext())
}

// Unwrap returns the underlying cause error
func (e *LSPError) Unwrap() error {
	return e.Cause
}

// formatContext formats the context information for error messages
func (e *LSPError) formatContext() string {
	if len(e.Context) == 0 {
		return ""
	}
	
	contextStr := ""
	for k, v := range e.Context {
		if contextStr != "" {
			contextStr += ", "
		}
		contextStr += fmt.Sprintf("%s=%v", k, v)
	}
	return fmt.Sprintf("[%s]", contextStr)
}

// ToJSONRPCError converts LSPError to jsonrpc2.Error
func (e *LSPError) ToJSONRPCError() *jsonrpc2.Error {
	var data *json.RawMessage
	if e.Data != nil {
		if raw, ok := e.Data.(*json.RawMessage); ok {
			data = raw
		}
	}
	return &jsonrpc2.Error{
		Code:    int64(e.Code),
		Message: e.Message,
		Data:    data,
	}
}

// WithContext adds context to the error
func (e *LSPError) WithContext(key string, value interface{}) *LSPError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// NewLSPError creates a new LSP error
func NewLSPError(code LSPErrorCode, message string) *LSPError {
	return &LSPError{
		Code:    code,
		Message: message,
		Context: make(map[string]interface{}),
	}
}

// NewLSPErrorWithCause creates a new LSP error with a causing error
func NewLSPErrorWithCause(code LSPErrorCode, message string, cause error) *LSPError {
	return &LSPError{
		Code:    code,
		Message: message,
		Cause:   cause,
		Context: make(map[string]interface{}),
	}
}

// NewLSPErrorWithData creates a new LSP error with additional data
func NewLSPErrorWithData(code LSPErrorCode, message string, data interface{}) *LSPError {
	return &LSPError{
		Code:    code,
		Message: message,
		Data:    data,
		Context: make(map[string]interface{}),
	}
}

// Common error creation functions
func NewParseError(message string, cause error) *LSPError {
	return NewLSPErrorWithCause(ErrorCodeParseError, message, cause)
}

func NewInvalidParamsError(message string, cause error) *LSPError {
	return NewLSPErrorWithCause(ErrorCodeInvalidParams, message, cause)
}

func NewMethodNotFoundError(method string) *LSPError {
	return NewLSPError(ErrorCodeMethodNotFound, fmt.Sprintf("method not found: %s", method)).
		WithContext("method", method)
}

func NewDocumentNotFoundError(uri string) *LSPError {
	return NewLSPError(ErrorCodeDocumentNotFound, fmt.Sprintf("document not found: %s", uri)).
		WithContext("uri", uri)
}

func NewInternalError(message string, cause error) *LSPError {
	return NewLSPErrorWithCause(ErrorCodeInternalError, message, cause)
}

// ErrorHandler provides a centralized way to handle errors in the LSP server
type ErrorHandler struct {
	server *MockLSPServer
}

// NewErrorHandler creates a new error handler
func NewErrorHandler(server *MockLSPServer) *ErrorHandler {
	return &ErrorHandler{server: server}
}

// HandleError processes an error and logs it appropriately
func (eh *ErrorHandler) HandleError(err error, operation string) {
	if err == nil {
		return
	}

	if lspErr, ok := err.(*LSPError); ok {
		// Log structured error with context
		if eh.server.structuredLogger != nil {
			logger := eh.server.structuredLogger.WithContext("operation", operation).WithContext("error_code", lspErr.Code.String())
			for k, v := range lspErr.Context {
				logger = logger.WithContext(k, v)
			}
			logger.Error("LSP operation failed: %s", lspErr.Message)
		} else {
			eh.server.logError("LSP operation failed [%s]: %v", operation, err)
		}
	} else {
		// Log generic error
		if eh.server.structuredLogger != nil {
			eh.server.structuredLogger.WithContext("operation", operation).Error("Operation failed: %v", err)
		} else {
			eh.server.logError("Operation failed [%s]: %v", operation, err)
		}
	}
}

// WrapError wraps a generic error into an LSPError with context
func (eh *ErrorHandler) WrapError(err error, code LSPErrorCode, message string, context map[string]interface{}) *LSPError {
	lspErr := NewLSPErrorWithCause(code, message, err)
	for k, v := range context {
		lspErr.WithContext(k, v)
	}
	return lspErr
}