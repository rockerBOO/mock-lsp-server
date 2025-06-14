package lsp

import (
	"errors"
	"testing"
)

func TestLSPErrorCode_String(t *testing.T) {
	testCases := []struct {
		code     LSPErrorCode
		expected string
	}{
		{ErrorCodeParseError, "ParseError"},
		{ErrorCodeInvalidRequest, "InvalidRequest"},
		{ErrorCodeMethodNotFound, "MethodNotFound"},
		{ErrorCodeInvalidParams, "InvalidParams"},
		{ErrorCodeInternalError, "InternalError"},
		{ErrorCodeDocumentNotFound, "DocumentNotFound"},
		{ErrorCodeInvalidDocument, "InvalidDocument"},
		{LSPErrorCode(9999), "UnknownError"}, // Unknown code
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			if tc.code.String() != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, tc.code.String())
			}
		})
	}
}

func TestNewLSPError(t *testing.T) {
	code := ErrorCodeInvalidParams
	message := "test error message"
	
	err := NewLSPError(code, message)
	
	if err.Code != code {
		t.Errorf("Expected code %v, got %v", code, err.Code)
	}
	
	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}
	
	if err.Context == nil {
		t.Error("Expected context to be initialized")
	}
}

func TestNewLSPErrorWithCause(t *testing.T) {
	code := ErrorCodeInternalError
	message := "test error with cause"
	cause := errors.New("underlying error")
	
	err := NewLSPErrorWithCause(code, message, cause)
	
	if err.Code != code {
		t.Errorf("Expected code %v, got %v", code, err.Code)
	}
	
	if err.Message != message {
		t.Errorf("Expected message %s, got %s", message, err.Message)
	}
	
	if err.Cause != cause {
		t.Errorf("Expected cause %v, got %v", cause, err.Cause)
	}
	
	// Test Unwrap method
	if err.Unwrap() != cause {
		t.Errorf("Expected Unwrap to return %v, got %v", cause, err.Unwrap())
	}
}

func TestLSPError_WithContext(t *testing.T) {
	err := NewLSPError(ErrorCodeInvalidParams, "test error")
	
	err.WithContext("method", "textDocument/completion")
	err.WithContext("uri", "file:///test.go")
	
	if len(err.Context) != 2 {
		t.Errorf("Expected 2 context entries, got %d", len(err.Context))
	}
	
	if err.Context["method"] != "textDocument/completion" {
		t.Errorf("Expected method context, got %v", err.Context["method"])
	}
	
	if err.Context["uri"] != "file:///test.go" {
		t.Errorf("Expected uri context, got %v", err.Context["uri"])
	}
}

func TestLSPError_Error(t *testing.T) {
	// Test error without cause
	err1 := NewLSPError(ErrorCodeInvalidParams, "invalid parameters")
	err1.WithContext("method", "test")
	
	errorStr1 := err1.Error()
	if errorStr1 == "" {
		t.Error("Error string should not be empty")
	}
	
	// Test error with cause
	cause := errors.New("underlying error")
	err2 := NewLSPErrorWithCause(ErrorCodeInternalError, "internal error", cause)
	err2.WithContext("operation", "parse")
	
	errorStr2 := err2.Error()
	if errorStr2 == "" {
		t.Error("Error string should not be empty")
	}
	
	// Should contain cause information
	if errorStr2 == errorStr1 {
		t.Error("Error with cause should be different from error without cause")
	}
}

func TestLSPError_ToJSONRPCError(t *testing.T) {
	lspErr := NewLSPError(ErrorCodeInvalidParams, "test error")
	lspErr.WithContext("method", "test")
	
	rpcErr := lspErr.ToJSONRPCError()
	
	if rpcErr == nil {
		t.Fatal("ToJSONRPCError returned nil")
	}
	
	if rpcErr.Code != int64(ErrorCodeInvalidParams) {
		t.Errorf("Expected code %d, got %d", int64(ErrorCodeInvalidParams), rpcErr.Code)
	}
	
	if rpcErr.Message != "test error" {
		t.Errorf("Expected message 'test error', got %s", rpcErr.Message)
	}
}

func TestCommonErrorCreationFunctions(t *testing.T) {
	// Test NewParseError
	parseErr := NewParseError("parse failed", errors.New("json error"))
	if parseErr.Code != ErrorCodeParseError {
		t.Errorf("Expected ParseError code, got %v", parseErr.Code)
	}
	
	// Test NewInvalidParamsError
	paramsErr := NewInvalidParamsError("invalid params", errors.New("unmarshal error"))
	if paramsErr.Code != ErrorCodeInvalidParams {
		t.Errorf("Expected InvalidParams code, got %v", paramsErr.Code)
	}
	
	// Test NewMethodNotFoundError
	methodErr := NewMethodNotFoundError("unknown/method")
	if methodErr.Code != ErrorCodeMethodNotFound {
		t.Errorf("Expected MethodNotFound code, got %v", methodErr.Code)
	}
	if methodErr.Context["method"] != "unknown/method" {
		t.Errorf("Expected method context, got %v", methodErr.Context["method"])
	}
	
	// Test NewDocumentNotFoundError
	docErr := NewDocumentNotFoundError("file:///missing.go")
	if docErr.Code != ErrorCodeDocumentNotFound {
		t.Errorf("Expected DocumentNotFound code, got %v", docErr.Code)
	}
	if docErr.Context["uri"] != "file:///missing.go" {
		t.Errorf("Expected uri context, got %v", docErr.Context["uri"])
	}
	
	// Test NewInternalError
	internalErr := NewInternalError("server error", errors.New("database error"))
	if internalErr.Code != ErrorCodeInternalError {
		t.Errorf("Expected InternalError code, got %v", internalErr.Code)
	}
}

func TestErrorHandler(t *testing.T) {
	// Create a test server
	server := createTestServer()
	errorHandler := NewErrorHandler(server)
	
	if errorHandler == nil {
		t.Fatal("NewErrorHandler returned nil")
	}
	
	if errorHandler.server != server {
		t.Error("ErrorHandler server reference is incorrect")
	}
	
	// Test HandleError with nil error (should not panic)
	errorHandler.HandleError(nil, "test_operation")
	
	// Test HandleError with LSPError
	lspErr := NewLSPError(ErrorCodeInvalidParams, "test error")
	lspErr.WithContext("method", "test")
	errorHandler.HandleError(lspErr, "test_operation")
	
	// Test HandleError with generic error
	genericErr := errors.New("generic error")
	errorHandler.HandleError(genericErr, "test_operation")
}

func TestErrorHandler_WrapError(t *testing.T) {
	server := createTestServer()
	errorHandler := NewErrorHandler(server)
	
	originalErr := errors.New("original error")
	context := map[string]interface{}{
		"method": "test",
		"uri":    "file:///test.go",
	}
	
	wrappedErr := errorHandler.WrapError(originalErr, ErrorCodeInternalError, "wrapped error", context)
	
	if wrappedErr.Code != ErrorCodeInternalError {
		t.Errorf("Expected InternalError code, got %v", wrappedErr.Code)
	}
	
	if wrappedErr.Message != "wrapped error" {
		t.Errorf("Expected 'wrapped error' message, got %s", wrappedErr.Message)
	}
	
	if wrappedErr.Cause != originalErr {
		t.Errorf("Expected original error as cause, got %v", wrappedErr.Cause)
	}
	
	if len(wrappedErr.Context) != 2 {
		t.Errorf("Expected 2 context entries, got %d", len(wrappedErr.Context))
	}
	
	if wrappedErr.Context["method"] != "test" {
		t.Errorf("Expected method context, got %v", wrappedErr.Context["method"])
	}
}

func TestLSPError_formatContext(t *testing.T) {
	// Test empty context
	err1 := NewLSPError(ErrorCodeInvalidParams, "test")
	if err1.formatContext() != "" {
		t.Errorf("Expected empty context format, got %s", err1.formatContext())
	}
	
	// Test single context
	err2 := NewLSPError(ErrorCodeInvalidParams, "test")
	err2.WithContext("method", "test")
	contextStr := err2.formatContext()
	if contextStr == "" {
		t.Error("Expected non-empty context format")
	}
	
	// Test multiple context entries
	err3 := NewLSPError(ErrorCodeInvalidParams, "test")
	err3.WithContext("method", "test")
	err3.WithContext("uri", "file:///test.go")
	multiContextStr := err3.formatContext()
	if multiContextStr == "" {
		t.Error("Expected non-empty multi-context format")
	}
	
	// Should be different from single context
	if multiContextStr == contextStr {
		t.Error("Multi-context should be different from single context")
	}
}