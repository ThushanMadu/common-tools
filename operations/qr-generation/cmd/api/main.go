// Copyright (c) 2026 WSO2 LLC. (https://www.wso2.com).
//
// WSO2 LLC. licenses this file to you under the Apache License,
// Version 2.0 (the "License"); you may not use this file except
// in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package main is the entry point for the QR code generation service.
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/wso2-open-operations/common-tools/operations/qr-generation/internal/config"
	"github.com/wso2-open-operations/common-tools/operations/qr-generation/internal/logger"
	"github.com/wso2-open-operations/common-tools/operations/qr-generation/internal/qr"
	transport "github.com/wso2-open-operations/common-tools/operations/qr-generation/internal/transport/http"
)

func main() {
	logger.InitLogger()
	logger.Logger.Debug("Starting QR generation service initialization")

	cfg := config.LoadConfig()
	logger.Logger.Debug("Configuration loaded",
		"port", cfg.Port,
		"read_timeout", cfg.ReadTimeout,
		"write_timeout", cfg.WriteTimeout,
		"max_body_size", cfg.MaxBodySize,
	)

	svc := qr.NewService(logger.Logger)
	logger.Logger.Debug("QR service initialized")

	h := transport.NewHandler(svc, logger.Logger, cfg.MaxBodySize)
	logger.Logger.Debug("HTTP handler initialized", "max_body_size", cfg.MaxBodySize)

	mux := http.NewServeMux()
	mux.HandleFunc("/generate", h.Generate)
	mux.HandleFunc("/health", h.HealthCheck)
	logger.Logger.Debug("HTTP routes registered", "endpoints", []string{"/generate", "/health"})

	// Configure HTTP server with timeouts and security settings
	srv := &http.Server{
		Addr:              fmt.Sprintf(":%s", cfg.Port),
		Handler:           mux,
		ReadTimeout:       cfg.ReadTimeout,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      cfg.WriteTimeout,
		IdleTimeout:       60 * time.Second,
	}
	logger.Logger.Debug("HTTP server configured",
		"addr", srv.Addr,
		"read_timeout", cfg.ReadTimeout,
		"write_timeout", cfg.WriteTimeout,
	)

	go func() {
		logger.Logger.Info("Starting server", "port", cfg.Port, "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Logger.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	logger.Logger.Info("Shutdown signal received", "signal", sig.String())
	logger.Logger.Debug("Initiating graceful shutdown", "timeout", cfg.ShutdownTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Logger.Error("Server forced to shutdown", "error", err, "timeout", cfg.ShutdownTimeout)
		os.Exit(1)
	}

	logger.Logger.Info("Server exited gracefully")
}
