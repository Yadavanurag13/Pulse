// services/user-service/internal/utils/logger/logger.go
package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a global SugaredLogger instance for convenient logging throughout the application.
var Logger *zap.SugaredLogger

// InitLogger initializes the global Zap logger based on the application environment.
func InitLogger(env string) {
	var config zap.Config
	if env == "production" {
		// Production configuration: JSON format, Info level by default
		config = zap.NewProductionConfig()
		config.Encoding = "json"
		config.Level.SetLevel(zap.InfoLevel)
	} else {
		// Development configuration: console format, Debug level by default, with colors
		config = zap.NewDevelopmentConfig()
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder // Add colors for dev
		config.Level.SetLevel(zap.DebugLevel) // More verbose logging in dev
	}

	// Direct output to standard streams
	config.OutputPaths = []string{"stdout"}
	config.ErrorOutputPaths = []string{"stderr"}

	// Build the logger instance
	l, err := config.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to build zap logger: %v", err))
	}

	// Assign to the global SugaredLogger variable for easy access
	Logger = l.Sugar()

	// Replace Zap's global logger with this configured one.
	zap.ReplaceGlobals(l)

	// Ensure all buffered logs are flushed when the application exits.
	defer func() {
		// Ignore common Windows error with stderr sync
		if err := l.Sync(); err != nil && err.Error() != "sync /dev/stderr: invalid argument" {
			fmt.Fprintf(os.Stderr, "failed to sync logger: %v\n", err)
		}
	}()
}