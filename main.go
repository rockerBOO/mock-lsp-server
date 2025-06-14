package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/sourcegraph/jsonrpc2"
	"io"
	"log"
	"os"
	"os/user"

	"mock-lsp-server/logging"
	"mock-lsp-server/lsp"
)

// func parseFlags() (config *Config, output string, err error) {
func loadConfig(progname string, args []string) (*MockLSPServerConfig, error) {
	flags := flag.NewFlagSet(progname, flag.ContinueOnError)

	var conf MockLSPServerConfig
	flags.StringVar(&conf.AppName, "appName", "mock-lsp-server", "set application name")
	flags.StringVar(&conf.LogDir, "log_dir", "", "set log directory")
	flags.StringVar(&conf.ConfigPath, "config", "", "set config file")
	flags.BoolVar(&conf.ShowInfo, "info", false, "set show info flag")

	err := flags.Parse(args)

	if err != nil {
		return nil, err
	}

	return &conf, nil
}

type MockLSPServerConfig struct {
	AppName    string
	LogDir     string
	ConfigPath string
	ShowInfo   bool
}

func main() {
	config, err := loadConfig(os.Args[0], os.Args[1:])

	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Configure logging
	logger, logManager, err := setupLogging(config.AppName, config.LogDir, config.ConfigPath, config.ShowInfo)

	if err != nil {
		log.Fatalf("Failed to setup logging: %v", err)
	}

	defer logManager.Close()

	logger.Println("Starting Mock LSP Server...")

	// Create structured logger for better logging
	structuredLogger := logManager.NewStructuredLogger().WithContext("component", "lsp-server")
	server := lsp.NewMockLSPServerWithStructuredLogger(structuredLogger, logger)

	// Create JSON-RPC connection using stdio
	handler := func(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) (any, error) {
		server.Handle(ctx, conn, req)
		return nil, nil
	}

	readWriteCloser := newStdioReadWriteCloser()
	ctx := context.Background()

	conn := jsonrpc2.NewConn(
		ctx,
		jsonrpc2.NewBufferedStream(readWriteCloser, jsonrpc2.VSCodeObjectCodec{}),
		jsonrpc2.HandlerWithError(handler),
		jsonrpc2.SetLogger(logger),
	)

	defer conn.Close()

	structuredLogger.Info("Mock LSP Server started, waiting for requests...")

	// Wait for the connection to close
	<-conn.DisconnectNotify()
	log.Println("Mock LSP Server stopped")
}

// stdioReadWriteCloser combines stdin and stdout into a single ReadWriteCloser
type stdioReadWriteCloser struct {
	io.Reader
	io.Writer
}

func (rw *stdioReadWriteCloser) Close() error {
	// For stdio, we typically don't want to close stdin/stdout
	// but we need to implement the interface
	return nil
}

func newStdioReadWriteCloser() io.ReadWriteCloser {
	return &stdioReadWriteCloser{
		Reader: os.Stdin,
		Writer: os.Stdout,
	}
}

func setupLogging(appName string, logDir, configPath string, showInfo bool) (*log.Logger, *logging.Manager, error) {
	u, err := user.Current()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get current user: %v", err)
	}

	// Create logging manager
	logManager := logging.NewManager(appName, u, true)

	// Get default config path if not specified
	if configPath == "" {
		configPath, err = logManager.GetDefaultConfigPath()
		if err != nil {
			log.Printf("Warning: failed to get default config path: %v", err)
			configPath = "" // Continue without config
		}
	}

	// Initialize logging system
	if err := logManager.Initialize(logDir, configPath); err != nil {
		return nil, nil, fmt.Errorf("failed to initialize logging: %v", err)
	}

	// Get logging information
	logInfo, err := logManager.GetInfo(logDir)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get log info: %v", err)
	}

	// Get the logger
	logger := logManager.GetLogger()

	// Show configuration info if requested
	if showInfo {
		printLogInfo(logInfo, logger)
	}

	return logger, logManager, nil
}

func printLogInfo(info *logging.LogInfo, logger *log.Logger) {
	logger.Printf("=== Logging Configuration ===\n")
	logger.Printf("App Name: %s\n", info.AppName)
	logger.Printf("Log Directory: %s\n", info.LogDirectory)
	logger.Printf("Log File Path: %s\n", info.LogFilePath)
	logger.Printf("Log File Name: %s\n", info.LogFileName)
	logger.Printf("Config Path: %s\n", info.ConfigPath)

	logger.Printf("\n=== Directory Resolution ===\n")
	if info.UsingCLIDir {
		logger.Printf("✓ Using CLI-specified directory\n")
	} else if info.UsingConfigDir {
		logger.Printf("✓ Using config file directory\n")
	} else {
		logger.Printf("✓ Using user-specific default directory\n")
	}

	// Check if config file exists
	if _, err := os.Stat(info.ConfigPath); err == nil {
		logger.Printf("✓ Config file exists: %s\n", info.ConfigPath)
	} else {
		logger.Printf("• Config file not found (using defaults): %s\n", info.ConfigPath)
	}
}
