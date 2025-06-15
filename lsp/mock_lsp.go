package lsp

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"reflect"

	"github.com/myleshyson/lsprotocol-go/protocol"
	"github.com/sourcegraph/jsonrpc2"
	"mock-lsp-server/logging"
)

// MockLSPServer implements the LSP server handlers
type MockLSPServer struct {
	documents        map[string]*protocol.TextDocumentItem
	logger           *log.Logger
	structuredLogger *logging.StructuredLogger
	errorHandler     *ErrorHandler
}

// NewMockLSPServer creates a new mock LSP server instance
func NewMockLSPServer(logger *log.Logger) *MockLSPServer {
	server := &MockLSPServer{
		documents: make(map[string]*protocol.TextDocumentItem),
		logger:    logger,
	}
	server.errorHandler = NewErrorHandler(server)
	return server
}

// NewMockLSPServerWithStructuredLogger creates a new mock LSP server with structured logging
func NewMockLSPServerWithStructuredLogger(structuredLogger *logging.StructuredLogger, fallbackLogger *log.Logger) *MockLSPServer {
	server := &MockLSPServer{
		documents:        make(map[string]*protocol.TextDocumentItem),
		logger:           fallbackLogger,
		structuredLogger: structuredLogger,
	}
	server.errorHandler = NewErrorHandler(server)
	return server
}

// logInfo logs an info message using structured logger if available, otherwise fallback
func (s *MockLSPServer) logInfo(format string, args ...interface{}) {
	if s.structuredLogger != nil {
		s.structuredLogger.Info(format, args...)
	} else {
		s.logger.Printf(format, args...)
	}
}

// logError logs an error message using structured logger if available, otherwise fallback
func (s *MockLSPServer) logError(format string, args ...interface{}) {
	if s.structuredLogger != nil {
		s.structuredLogger.Error(format, args...)
	} else {
		s.logger.Printf("ERROR: "+format, args...)
	}
}

// logDebug logs a debug message using structured logger if available, otherwise fallback
func (s *MockLSPServer) logDebug(format string, args ...interface{}) {
	if s.structuredLogger != nil {
		s.structuredLogger.Debug(format, args...)
	} else {
		s.logger.Printf("DEBUG: "+format, args...)
	}
}

// logWarning logs a warning message using structured logger if available, otherwise fallback
func (s *MockLSPServer) logWarning(format string, args ...interface{}) {
	if s.structuredLogger != nil {
		s.structuredLogger.Warning(format, args...)
	} else {
		s.logger.Printf("WARNING: "+format, args...)
	}
}

// Handle processes incoming JSON-RPC requests
func (s *MockLSPServer) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	switch req.Method {
	case "initialize":
		s.handleInitialize(ctx, conn, req)
	case "initialized":
		s.handleInitialized(ctx, conn, req)
	case "textDocument/didOpen":
		s.handleTextDocumentDidOpen(ctx, conn, req)
	case "textDocument/didChange":
		s.handleTextDocumentDidChange(ctx, conn, req)
	case "textDocument/didSave":
		s.handleTextDocumentDidSave(ctx, conn, req)
	case "textDocument/didClose":
		s.handleTextDocumentDidClose(ctx, conn, req)
	case "textDocument/completion":
		s.handleCompletion(ctx, conn, req)
	case "textDocument/hover":
		s.handleHover(ctx, conn, req)
	case "textDocument/definition":
		s.handleDefinition(ctx, conn, req)
	case "textDocument/references":
		s.handleReferences(ctx, conn, req)
	case "textDocument/documentSymbol":
		s.handleDocumentSymbol(ctx, conn, req)
	case "shutdown":
		s.handleShutdown(ctx, conn, req)
	case "exit":
		s.handleExit(ctx, conn, req)
	default:
		// Create structured error for unsupported method
		lspErr := NewMethodNotFoundError(req.Method)
		if err := conn.ReplyWithError(ctx, req.ID, lspErr.ToJSONRPCError()); err != nil {
			// Handle reply error with context
			replyErr := s.errorHandler.WrapError(err, ErrorCodeInternalError, "Failed to send method not found error", map[string]interface{}{
				"method":     req.Method,
				"request_id": req.ID,
			})
			s.errorHandler.HandleError(replyErr, "handle_unsupported_method")
		}
	}
}

// handleInitialize processes the initialize request
func (s *MockLSPServer) handleInitialize(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.InitializeParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		lspErr := NewInvalidParamsError("failed to parse initialize params", err)
		lspErr.WithContext("method", "initialize")
		if replyErr := conn.ReplyWithError(ctx, req.ID, lspErr.ToJSONRPCError()); replyErr != nil {
			s.errorHandler.HandleError(replyErr, "initialize_send_error")
		}
		s.errorHandler.HandleError(lspErr, "initialize_parse_params")
		return
	}

	s.logInfo("Initialize request from client with root URI: %+v", params.RootUri)

	// textDocumentSyncChange := protocol.TextDocumentSyncKind(0)

	textDocumentSync := protocol.Or2[protocol.TextDocumentSyncOptions, protocol.TextDocumentSyncKind]{Value: protocol.TextDocumentSyncKind(0)}

	completionProvider := protocol.CompletionOptions{TriggerCharacters: []string{".", ":"}}
	hoverProvider := protocol.Or2[bool, protocol.HoverOptions]{Value: true}
	definitionProvider := protocol.Or2[bool, protocol.DefinitionOptions]{Value: true}
	referencesProvider := protocol.Or2[bool, protocol.ReferenceOptions]{Value: true}
	documentSymbolProvider := protocol.Or2[bool, protocol.DocumentSymbolOptions]{Value: true}

	// Mock server capabilities
	result := protocol.InitializeResult{
		Capabilities: protocol.ServerCapabilities{
			TextDocumentSync:       &textDocumentSync,
			CompletionProvider:     &completionProvider,
			HoverProvider:          &hoverProvider,
			DefinitionProvider:     &definitionProvider,
			ReferencesProvider:     &referencesProvider,
			DocumentSymbolProvider: &documentSymbolProvider,
		},
		ServerInfo: &protocol.ServerInfo{
			Name:    "Mock LSP Server",
			Version: "1.0.0",
		},
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		replyErr := s.errorHandler.WrapError(err, ErrorCodeInternalError, "Failed to send initialize response", map[string]interface{}{
			"method":     "initialize",
			"request_id": req.ID,
		})
		s.errorHandler.HandleError(replyErr, "initialize_send_response")
	}
}

// handleInitialized processes the initialized notification
func (s *MockLSPServer) handleInitialized(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) {
	s.logInfo("Client initialized")
}

// handleTextDocumentDidOpen processes textDocument/didOpen notifications
func (s *MockLSPServer) handleTextDocumentDidOpen(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.DidOpenTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		lspErr := NewInvalidParamsError("failed to parse textDocument/didOpen params", err)
		lspErr.WithContext("method", "textDocument/didOpen")
		s.errorHandler.HandleError(lspErr, "didOpen_parse_params")
		return
	}

	s.documents[string(params.TextDocument.Uri)] = &params.TextDocument
	s.logger.Printf("Opened document: %s", params.TextDocument.Uri)

	// Send mock diagnostics
	s.sendMockDiagnostics(ctx, conn, string(params.TextDocument.Uri))
}

// handleTextDocumentDidChange processes textDocument/didChange notifications
func (s *MockLSPServer) handleTextDocumentDidChange(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.DidChangeTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		s.logger.Printf("Failed to parse didChange params: %v", err)
		return
	}

	uri := string(params.TextDocument.Uri)
	if doc, exists := s.documents[uri]; exists {
		// Update document version
		doc.Version = params.TextDocument.Version

		// Apply content changes
		for _, change := range params.ContentChanges {
			// Use reflection to get the actual value from the Or2 union type
			changeValue := reflect.ValueOf(change)

			// Get the Value field from the Or2 struct
			valueField := changeValue.FieldByName("Value")
			if !valueField.IsValid() {
				s.logger.Printf("Or2 union type doesn't have Value field")
				continue
			}

			// Get the actual underlying value
			actualValue := valueField.Interface()

			// Type switch on the actual concrete type
			switch v := actualValue.(type) {
			case protocol.TextDocumentContentChangePartial:
				// Partial document change with range
				s.logger.Printf("Partial document update for %s at range %v", uri, v.Range)
				s.logger.Printf("Replacing text in range with: %q", v.Text)
				// In a real implementation, apply the range-based change
				// For this mock, we'll just note the change

			case protocol.TextDocumentContentChangeWholeDocument:
				// Whole document change
				doc.Text = v.Text
				s.logger.Printf("Full document update for %s", uri)

			default:
				s.logger.Printf("Unknown content change type: %T", v)
			}
		}

		s.logger.Printf("Document changed: %s (version %d)", uri, params.TextDocument.Version)

		// Send updated diagnostics after document change
		s.sendMockDiagnostics(ctx, conn, uri)
	}
}

// handleTextDocumentDidSave processes textDocument/didSave notifications
func (s *MockLSPServer) handleTextDocumentDidSave(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.DidSaveTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		s.logger.Printf("Failed to parse didSave params: %v", err)
		return
	}

	s.logger.Printf("Document saved: %s", params.TextDocument.Uri)
}

// handleTextDocumentDidClose processes textDocument/didClose notifications
func (s *MockLSPServer) handleTextDocumentDidClose(_ context.Context, _ *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.DidCloseTextDocumentParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		s.logger.Printf("Failed to parse didClose params: %v", err)
		return
	}

	delete(s.documents, string(params.TextDocument.Uri))
	s.logger.Printf("Closed document: %s", params.TextDocument.Uri)
}

// handleCompletion processes textDocument/completion requests
func (s *MockLSPServer) handleCompletion(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.CompletionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		if replyErr := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "failed to parse completion params",
		}); replyErr != nil {
			s.logger.Printf("Failed to send completion error: %v", replyErr)
		}
		return
	}

	// Mock completion items
	kind1 := protocol.CompletionItemKind(protocol.CompletionItemKindFunction)
	kind2 := protocol.CompletionItemKind(protocol.CompletionItemKindVariable)
	kind3 := protocol.CompletionItemKind(protocol.CompletionItemKindClass)

	items := []protocol.CompletionItem{
		{
			Label:  "mockFunction",
			Kind:   &kind1,
			Detail: "Mock function completion",
			Documentation: &protocol.Or2[string, protocol.MarkupContent]{
				Value: &protocol.MarkupContent{
					Kind:  protocol.MarkupKindMarkdown,
					Value: "This is a mock function completion",
				},
			},
			InsertText: "mockFunction()",
		},
		{
			Label:  "mockVariable",
			Kind:   &kind2,
			Detail: "Mock variable completion",
			Documentation: &protocol.Or2[string, protocol.MarkupContent]{
				Value: "This is a mock variable",
			},
		},
		{
			Label:      "mockClass",
			Kind:       &kind3,
			Detail:     "Mock class completion",
			InsertText: "MockClass",
		},
	}

	result := protocol.CompletionList{
		IsIncomplete: false,
		Items:        items,
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		s.logger.Printf("Failed to send completion response: %v", err)
	}
}

// handleHover processes textDocument/hover requests
func (s *MockLSPServer) handleHover(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.HoverParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		if replyErr := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "failed to parse hover params",
		}); replyErr != nil {
			s.logger.Printf("Failed to send hover error: %v", replyErr)
		}
		return
	}

	// Mock hover information
	result := protocol.Hover{
		Contents: protocol.Or3[protocol.MarkupContent, protocol.MarkedString, []protocol.MarkedString]{
			Value: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: "**Mock Hover Information**\n\nThis is mock hover content for testing purposes.",
			},
		},
		Range: &protocol.Range{
			Start: params.Position,
			End: protocol.Position{
				Line:      params.Position.Line,
				Character: params.Position.Character + 10, // Mock word length
			},
		},
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		s.logger.Printf("Failed to send hover response: %v", err)
	}
}

// handleDefinition processes textDocument/definition requests
func (s *MockLSPServer) handleDefinition(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.DefinitionParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		if replyErr := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "failed to parse definition params",
		}); replyErr != nil {
			s.logger.Printf("Failed to send definition error: %v", replyErr)
		}
		return
	}

	// Mock definition location
	result := []protocol.Location{
		{
			Uri: params.TextDocument.Uri,
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 10},
			},
		},
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		s.logger.Printf("Failed to send definition response: %v", err)
	}
}

// handleReferences processes textDocument/references requests
func (s *MockLSPServer) handleReferences(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.ReferenceParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		if replyErr := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "failed to parse references params",
		}); replyErr != nil {
			s.logger.Printf("Failed to send references error: %v", replyErr)
		}
		return
	}

	// Mock references
	result := []protocol.Location{
		{
			Uri: params.TextDocument.Uri,
			Range: protocol.Range{
				Start: protocol.Position{Line: 5, Character: 10},
				End:   protocol.Position{Line: 5, Character: 20},
			},
		},
		{
			Uri: params.TextDocument.Uri,
			Range: protocol.Range{
				Start: protocol.Position{Line: 10, Character: 5},
				End:   protocol.Position{Line: 10, Character: 15},
			},
		},
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		s.logger.Printf("Failed to send references response: %v", err)
	}
}

// handleDocumentSymbol processes textDocument/documentSymbol requests
func (s *MockLSPServer) handleDocumentSymbol(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	var params protocol.DocumentSymbolParams
	if err := json.Unmarshal(*req.Params, &params); err != nil {
		if replyErr := conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    jsonrpc2.CodeInvalidParams,
			Message: "failed to parse document symbol params",
		}); replyErr != nil {
			s.logger.Printf("Failed to send document symbol error: %v", replyErr)
		}
		return
	}

	// Mock document symbols
	result := []protocol.DocumentSymbol{
		{
			Name:   "MockClass",
			Kind:   protocol.SymbolKindClass,
			Detail: "Mock class symbol",
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 20, Character: 0},
			},
			SelectionRange: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 6},
				End:   protocol.Position{Line: 0, Character: 15},
			},
			Children: []protocol.DocumentSymbol{
				{
					Name: "mockMethod",
					Kind: protocol.SymbolKindMethod,
					Range: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 4},
						End:   protocol.Position{Line: 10, Character: 4},
					},
					SelectionRange: protocol.Range{
						Start: protocol.Position{Line: 5, Character: 4},
						End:   protocol.Position{Line: 5, Character: 14},
					},
				},
			},
		},
	}

	if err := conn.Reply(ctx, req.ID, result); err != nil {
		s.logger.Printf("Failed to send document symbol response: %v", err)
	}
}

// handleShutdown processes shutdown requests
func (s *MockLSPServer) handleShutdown(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	s.logger.Println("Shutdown request received")
	if err := conn.Reply(ctx, req.ID, nil); err != nil {
		s.logger.Printf("Failed to send shutdown response: %v", err)
	}
}

// handleExit processes exit notifications
func (s *MockLSPServer) handleExit(_ context.Context, _ *jsonrpc2.Conn, _ *jsonrpc2.Request) {
	s.logger.Println("Exit notification received")
	os.Exit(0)
}

// sendMockDiagnostics sends mock diagnostic information for a document
func (s *MockLSPServer) sendMockDiagnostics(ctx context.Context, conn *jsonrpc2.Conn, uri string) {
	severity1 := protocol.DiagnosticSeverity(protocol.DiagnosticSeverityWarning)
	severity2 := protocol.DiagnosticSeverity(protocol.DiagnosticSeverityInformation)

	diagnostics := []protocol.Diagnostic{
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 1, Character: 0},
				End:   protocol.Position{Line: 1, Character: 10},
			},
			Severity: &severity1,
			Message:  "This is a mock warning",
			Source:   "mock-lsp",
		},
		{
			Range: protocol.Range{
				Start: protocol.Position{Line: 5, Character: 15},
				End:   protocol.Position{Line: 5, Character: 25},
			},
			Severity: &severity2,
			Message:  "This is mock info",
			Source:   "mock-lsp",
		},
	}

	params := protocol.PublishDiagnosticsParams{
		Uri:         protocol.DocumentUri(uri),
		Diagnostics: diagnostics,
	}

	if err := conn.Notify(ctx, "textDocument/publishDiagnostics", params); err != nil {
		s.logger.Printf("Failed to send diagnostics notification: %v", err)
	}
}
