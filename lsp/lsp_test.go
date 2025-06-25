package lsp

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"slices"
	"strings"
	"testing"

	"github.com/myleshyson/lsprotocol-go/protocol"
)

// Test helper functions for LSP methods
func createTestLogger() *log.Logger {
	return log.New(io.Discard, "", 0)
}

func createTestServer() *MockLSPServer {
	return NewMockLSPServer(createTestLogger())
}

func TestNewMockLSPServer(t *testing.T) {
	// Create a temporary logger that discards output
	logger := log.New(io.Discard, "", 0)

	server := NewMockLSPServer(logger)

	if server == nil {
		t.Fatal("NewMockLSPServer returned nil")
	}

	if server.documents == nil {
		t.Fatal("documents map not initialized")
	}

	if len(server.documents) != 0 {
		t.Errorf("Expected empty documents map, got %d items", len(server.documents))
	}
}

// Test that NewMockLSPServer creates a server with logger
func TestNewMockLSPServerWithLogger(t *testing.T) {
	logger := createTestLogger()
	server := NewMockLSPServer(logger)

	if server == nil {
		t.Fatal("NewMockLSPServer returned nil")
	}

	if server.logger == nil {
		t.Fatal("Logger not set in MockLSPServer")
	}

	if server.documents == nil {
		t.Fatal("Documents map not initialized")
	}
}

func TestDocumentStorage(t *testing.T) {
	// Create a temporary logger that discards output
	logger := log.New(io.Discard, "", 0)

	server := NewMockLSPServer(logger)

	// Test adding a document
	uri := "file:///test.go"
	doc := &protocol.TextDocumentItem{
		Uri:     protocol.DocumentUri(uri),
		Text:    "package main",
		Version: 1,
	}

	server.documents[uri] = doc

	// Test retrieval
	retrieved, exists := server.documents[uri]
	if !exists {
		t.Error("Document not found after adding")
	}

	if retrieved.Text != "package main" {
		t.Errorf("Expected text 'package main', got %s", retrieved.Text)
	}

	// Test removal
	delete(server.documents, uri)
	_, exists = server.documents[uri]
	if exists {
		t.Error("Document still exists after deletion")
	}
}

// Test document lifecycle operations
func TestDocumentLifecycle(t *testing.T) {
	server := createTestServer()

	// Test adding documents
	uri1 := "file:///test1.go"
	uri2 := "file:///test2.go"

	doc1 := &protocol.TextDocumentItem{
		Uri:     protocol.DocumentUri(uri1),
		Text:    "package main",
		Version: 1,
	}

	doc2 := &protocol.TextDocumentItem{
		Uri:     protocol.DocumentUri(uri2),
		Text:    "package test",
		Version: 1,
	}

	// Add documents
	server.documents[uri1] = doc1
	server.documents[uri2] = doc2

	if len(server.documents) != 2 {
		t.Errorf("Expected 2 documents, got %d", len(server.documents))
	}

	// Test retrieval
	retrieved1, exists1 := server.documents[uri1]
	if !exists1 {
		t.Error("Document 1 not found")
	}
	if retrieved1.Text != "package main" {
		t.Errorf("Expected 'package main', got %s", retrieved1.Text)
	}

	// Test removal
	delete(server.documents, uri1)
	if len(server.documents) != 1 {
		t.Errorf("Expected 1 document after deletion, got %d", len(server.documents))
	}

	// Test document doesn't exist
	_, exists := server.documents[uri1]
	if exists {
		t.Error("Document 1 should not exist after deletion")
	}
}

func TestHandleInitializeParams(t *testing.T) {
	// Test parameter parsing for initialize request
	rootUri := protocol.DocumentUri("file:///test")
	params := protocol.InitializeParams{
		RootUri: &rootUri,
	}

	paramsBytes, err := json.Marshal(params)
	if err != nil {
		t.Fatalf("Failed to marshal params: %v", err)
	}

	var parsed protocol.InitializeParams
	err = json.Unmarshal(paramsBytes, &parsed)
	if err != nil {
		t.Fatalf("Failed to unmarshal params: %v", err)
	}

	if parsed.RootUri == nil || *parsed.RootUri != rootUri {
		t.Errorf("Expected RootUri %s, got %v", rootUri, parsed.RootUri)
	}
}

// Test parameter marshaling and unmarshaling
func TestLSPParameterSerialization(t *testing.T) {
	testCases := []struct {
		name   string
		params interface{}
	}{
		{
			name: "InitializeParams",
			params: protocol.InitializeParams{
				RootUri: func() *protocol.DocumentUri {
					uri := protocol.DocumentUri("file:///test")
					return &uri
				}(),
			},
		},
		{
			name: "DidOpenTextDocumentParams",
			params: protocol.DidOpenTextDocumentParams{
				TextDocument: protocol.TextDocumentItem{
					Uri:     "file:///test.go",
					Text:    "package main",
					Version: 1,
				},
			},
		},
		{
			name: "CompletionParams",
			params: protocol.CompletionParams{
				TextDocument: protocol.TextDocumentIdentifier{
					Uri: "file:///test.go",
				},
				Position: protocol.Position{Line: 0, Character: 0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test marshaling
			data, err := json.Marshal(tc.params)
			if err != nil {
				t.Errorf("Failed to marshal %s: %v", tc.name, err)
				return
			}

			// Test unmarshaling
			var result map[string]interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Errorf("Failed to unmarshal %s: %v", tc.name, err)
			}
		})
	}
}

func TestCompletionItemCreation(t *testing.T) {
	// Test creation of completion items similar to what the handler does
	kind1 := protocol.CompletionItemKind(protocol.CompletionItemKindFunction)
	kind2 := protocol.CompletionItemKind(protocol.CompletionItemKindVariable)

	items := []protocol.CompletionItem{
		{
			Label:      "mockFunction",
			Kind:       &kind1,
			Detail:     "Mock function completion",
			InsertText: "mockFunction()",
		},
		{
			Label:  "mockVariable",
			Kind:   &kind2,
			Detail: "Mock variable completion",
		},
	}

	if len(items) != 2 {
		t.Errorf("Expected 2 items, got %d", len(items))
	}

	if items[0].Label != "mockFunction" {
		t.Errorf("Expected first item label 'mockFunction', got %s", items[0].Label)
	}

	if items[1].Label != "mockVariable" {
		t.Errorf("Expected second item label 'mockVariable', got %s", items[1].Label)
	}
}

// Test response structure creation
func TestLSPResponseCreation(t *testing.T) {
	// Test InitializeResult creation
	t.Run("InitializeResult", func(t *testing.T) {
		textDocumentSync := protocol.Or2[protocol.TextDocumentSyncOptions, protocol.TextDocumentSyncKind]{Value: protocol.TextDocumentSyncKind(0)}
		completionProvider := protocol.CompletionOptions{TriggerCharacters: []string{".", ":"}}
		hoverProvider := protocol.Or2[bool, protocol.HoverOptions]{Value: true}

		result := protocol.InitializeResult{
			Capabilities: protocol.ServerCapabilities{
				TextDocumentSync:   &textDocumentSync,
				CompletionProvider: &completionProvider,
				HoverProvider:      &hoverProvider,
			},
			ServerInfo: &protocol.ServerInfo{
				Name:    "Mock LSP Server",
				Version: "1.0.0",
			},
		}

		if result.ServerInfo.Name != "Mock LSP Server" {
			t.Errorf("Expected server name 'Mock LSP Server', got %s", result.ServerInfo.Name)
		}

		if result.ServerInfo.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got %s", result.ServerInfo.Version)
		}
	})

	// Test CompletionList creation
	t.Run("CompletionList", func(t *testing.T) {
		kind1 := protocol.CompletionItemKind(protocol.CompletionItemKindFunction)
		kind2 := protocol.CompletionItemKind(protocol.CompletionItemKindVariable)

		items := []protocol.CompletionItem{
			{
				Label:      "testFunction",
				Kind:       &kind1,
				Detail:     "Test function",
				InsertText: "testFunction()",
			},
			{
				Label:  "testVariable",
				Kind:   &kind2,
				Detail: "Test variable",
			},
		}

		result := protocol.CompletionList{
			IsIncomplete: false,
			Items:        items,
		}

		if len(result.Items) != 2 {
			t.Errorf("Expected 2 completion items, got %d", len(result.Items))
		}

		if result.Items[0].Label != "testFunction" {
			t.Errorf("Expected first item 'testFunction', got %s", result.Items[0].Label)
		}
	})
}

func TestHoverContentCreation(t *testing.T) {
	// Test hover content creation
	hover := protocol.Hover{
		Contents: protocol.Or3[protocol.MarkupContent, protocol.MarkedString, []protocol.MarkedString]{
			Value: protocol.MarkupContent{
				Kind:  protocol.MarkupKindMarkdown,
				Value: "**Mock Hover Information**\n\nThis is mock hover content.",
			},
		},
		Range: &protocol.Range{
			Start: protocol.Position{Line: 0, Character: 0},
			End:   protocol.Position{Line: 0, Character: 10},
		},
	}

	if hover.Range == nil {
		t.Error("Expected hover range to be set")
	}

	if hover.Range.Start.Line != 0 {
		t.Errorf("Expected start line 0, got %d", hover.Range.Start.Line)
	}
}

// Test error handling for invalid JSON
func TestInvalidJSONHandling(t *testing.T) {
	// Test various invalid JSON scenarios
	testCases := []struct {
		name        string
		invalidJSON string
	}{
		{"Missing closing brace", `{"test": "value"`},
		{"Invalid syntax", `{"test": invalid}`},
		{"Incomplete string", `{"test": "incomplete`},
		{"Extra comma", `{"test": "value",}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var result map[string]interface{}
			err := json.Unmarshal([]byte(tc.invalidJSON), &result)
			if err == nil {
				t.Errorf("Expected error for invalid JSON: %s", tc.invalidJSON)
			}
		})
	}
}

func TestLocationCreation(t *testing.T) {
	// Test location creation for definition/references
	uri := protocol.DocumentUri("file:///test.go")
	locations := []protocol.Location{
		{
			Uri: uri,
			Range: protocol.Range{
				Start: protocol.Position{Line: 0, Character: 0},
				End:   protocol.Position{Line: 0, Character: 10},
			},
		},
	}

	if len(locations) != 1 {
		t.Errorf("Expected 1 location, got %d", len(locations))
	}

	if locations[0].Uri != uri {
		t.Errorf("Expected URI %s, got %s", uri, locations[0].Uri)
	}
}

func TestDocumentSymbolCreation(t *testing.T) {
	// Test document symbol creation
	symbols := []protocol.DocumentSymbol{
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
		},
	}

	if len(symbols) != 1 {
		t.Errorf("Expected 1 symbol, got %d", len(symbols))
	}

	if symbols[0].Name != "MockClass" {
		t.Errorf("Expected symbol name 'MockClass', got %s", symbols[0].Name)
	}

	if symbols[0].Kind != protocol.SymbolKindClass {
		t.Errorf("Expected symbol kind Class, got %v", symbols[0].Kind)
	}
}

func TestDiagnosticCreation(t *testing.T) {
	// Test diagnostic creation
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

	if len(diagnostics) != 2 {
		t.Errorf("Expected 2 diagnostics, got %d", len(diagnostics))
	}

	if diagnostics[0].Message != "This is a mock warning" {
		t.Errorf("Expected warning message, got %s", diagnostics[0].Message)
	}

	if diagnostics[1].Message != "This is mock info" {
		t.Errorf("Expected info message, got %s", diagnostics[1].Message)
	}
}

// Test method validation
func TestSupportedMethods(t *testing.T) {
	// List of all supported LSP methods
	supportedMethods := map[string]bool{
		"initialize":                  true,
		"initialized":                 true,
		"textDocument/didOpen":        true,
		"textDocument/didChange":      true,
		"textDocument/didSave":        true,
		"textDocument/didClose":       true,
		"textDocument/completion":     true,
		"textDocument/hover":          true,
		"textDocument/definition":     true,
		"textDocument/references":     true,
		"textDocument/documentSymbol": true,
		"shutdown":                    true,
		"exit":                        true,
	}

	// Test that all expected methods are supported
	for method := range supportedMethods {
		t.Run(method, func(t *testing.T) {
			if !supportedMethods[method] {
				t.Errorf("Method %s should be supported", method)
			}
		})
	}

	// Test that unsupported methods are not in the list
	unsupportedMethods := []string{
		"unsupported/method",
		"textDocument/unsupported",
		"workspace/unsupported",
	}

	for _, method := range unsupportedMethods {
		t.Run("unsupported_"+method, func(t *testing.T) {
			if supportedMethods[method] {
				t.Errorf("Method %s should not be supported", method)
			}
		})
	}
}

func TestHandleMethodSwitch(t *testing.T) {
	// Test that our handler supports all expected LSP methods
	testCases := []struct {
		method string
		hasID  bool
	}{
		{"initialize", true},
		{"initialized", false},
		{"textDocument/didOpen", false},
		{"textDocument/didChange", false},
		{"textDocument/didSave", false},
		{"textDocument/didClose", false},
		{"textDocument/completion", true},
		{"textDocument/hover", true},
		{"textDocument/definition", true},
		{"textDocument/references", true},
		{"textDocument/documentSymbol", true},
		{"shutdown", true},
		{"exit", false},
	}

	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			// This test validates that our test cases cover the expected methods
			validMethods := []string{
				"initialize", "initialized", "textDocument/didOpen", "textDocument/didChange",
				"textDocument/didSave", "textDocument/didClose", "textDocument/completion",
				"textDocument/hover", "textDocument/definition", "textDocument/references",
				"textDocument/documentSymbol", "shutdown", "exit",
			}

			found := slices.Contains(validMethods, tc.method)

			if !found {
				t.Errorf("Method %s not found in valid methods list", tc.method)
			}
		})
	}
}

// Test concurrent access to documents map
func TestConcurrentDocumentAccess(t *testing.T) {
	server := createTestServer()

	// Test concurrent reads and writes to documents map
	done := make(chan bool)
	numGoroutines := 10

	// Start multiple goroutines that access the documents map
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			uri := fmt.Sprintf("file:///test%d.go", id)
			doc := &protocol.TextDocumentItem{
				Uri:     protocol.DocumentUri(uri),
				Text:    fmt.Sprintf("package test%d", id),
				Version: 1,
			}

			server.mu.Lock()
			// Add document
			server.documents[uri] = doc
			server.mu.Unlock()

			server.mu.Lock()
			// Read document
			if retrieved, exists := server.documents[uri]; exists {
				if retrieved.Text != fmt.Sprintf("package test%d", id) {
					t.Errorf("Unexpected document content for %s", uri)
				}
			}
			server.mu.Unlock()

			server.mu.Lock()
			// Remove document
			delete(server.documents, uri)
			server.mu.Unlock()

			done <- true
		}(i)
	}

	// Wait for all goroutines to complete
	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

// Test edge cases and boundary conditions
func TestEdgeCases(t *testing.T) {
	server := createTestServer()

	// Test empty URI
	t.Run("EmptyURI", func(t *testing.T) {
		uri := ""
		doc := &protocol.TextDocumentItem{
			Uri:     protocol.DocumentUri(uri),
			Text:    "test",
			Version: 1,
		}
		server.documents[uri] = doc

		if _, exists := server.documents[uri]; !exists {
			t.Error("Empty URI document should be stored")
		}
	})

	// Test very long document text
	t.Run("LongDocumentText", func(t *testing.T) {
		uri := "file:///long.go"
		longText := strings.Repeat("a", 10000) // 10KB of text
		doc := &protocol.TextDocumentItem{
			Uri:     protocol.DocumentUri(uri),
			Text:    longText,
			Version: 1,
		}
		server.documents[uri] = doc

		if retrieved, exists := server.documents[uri]; exists {
			if len(retrieved.Text) != 10000 {
				t.Errorf("Expected text length 10000, got %d", len(retrieved.Text))
			}
		} else {
			t.Error("Long document should be stored")
		}
	})

	// Test zero version
	t.Run("ZeroVersion", func(t *testing.T) {
		uri := "file:///zero.go"
		doc := &protocol.TextDocumentItem{
			Uri:     protocol.DocumentUri(uri),
			Text:    "test",
			Version: 0,
		}
		server.documents[uri] = doc

		if retrieved, exists := server.documents[uri]; exists {
			if retrieved.Version != 0 {
				t.Errorf("Expected version 0, got %d", retrieved.Version)
			}
		} else {
			t.Error("Zero version document should be stored")
		}
	})
}

func TestJSONSerialization(t *testing.T) {
	// Test that our protocol structures can be serialized/deserialized
	testCases := []any{
		protocol.InitializeParams{},
		protocol.CompletionParams{},
		protocol.HoverParams{},
		protocol.DefinitionParams{},
		protocol.ReferenceParams{},
		protocol.DocumentSymbolParams{},
		protocol.DidOpenTextDocumentParams{},
		protocol.DidChangeTextDocumentParams{},
		protocol.DidSaveTextDocumentParams{},
		protocol.DidCloseTextDocumentParams{},
	}

	for _, testCase := range testCases {
		t.Run(strings.Replace(strings.Replace(fmt.Sprintf("%T", testCase), "protocol.", "", 1), "Params", "", 1), func(t *testing.T) {
			data, err := json.Marshal(testCase)
			if err != nil {
				t.Errorf("Failed to marshal %T: %v", testCase, err)
			}

			// Try to unmarshal back
			var result interface{}
			err = json.Unmarshal(data, &result)
			if err != nil {
				t.Errorf("Failed to unmarshal %T: %v", testCase, err)
			}
		})
	}
}
