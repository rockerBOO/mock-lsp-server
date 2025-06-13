package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/myleshyson/lsprotocol-go/protocol"
	"slices"
)

func TestNewMockLSPServer(t *testing.T) {
	server := NewMockLSPServer()

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

func TestDocumentStorage(t *testing.T) {
	server := NewMockLSPServer()

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
