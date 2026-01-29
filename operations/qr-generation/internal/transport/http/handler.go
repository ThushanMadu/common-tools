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

// Package http provides HTTP transport layer for the QR code generation service.
package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/wso2-open-operations/common-tools/operations/qr-generation/internal/qr"
)

type Handler struct {
	svc         qr.Service
	logger      *slog.Logger
	maxBodySize int64
}

// NewHandler creates a new HTTP handler for QR code generation.
func NewHandler(svc qr.Service, logger *slog.Logger, maxBodySize int64) *Handler {
	return &Handler{
		svc:         svc,
		logger:      logger,
		maxBodySize: maxBodySize,
	}
}

// Generate handles POST /generate?size={pixels} requests to create QR codes.
// Accepts raw text/URL in body, returns PNG image.
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	// Log incoming request with metadata
	h.logger.Debug("Received QR generation request",
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
		"user_agent", r.UserAgent(),
		"content_length", r.ContentLength,
	)

	// Only accept POST requests
	if r.Method != http.MethodPost {
		h.logger.Warn("Method not allowed",
			"method", r.Method,
			"expected", http.MethodPost,
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Enforce maximum request body size to prevent DoS attacks
	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodySize)
	h.logger.Debug("Reading request body", "max_size", h.maxBodySize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("failed to read request body", "error", err, "remote_addr", r.RemoteAddr)
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			h.logger.Warn("Request body too large",
				"max_allowed", h.maxBodySize,
				"remote_addr", r.RemoteAddr,
			)
			http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
			return
		}
		http.Error(w, "Failed to read request body", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("Request body read successfully", "body_size", len(body))

	if len(body) == 0 {
		h.logger.Warn("Empty request body received", "remote_addr", r.RemoteAddr)
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	const maxSize = 2048
	const defaultSize = 256
	size := defaultSize
	sizeStr := r.URL.Query().Get("size")

	if sizeStr != "" {
		h.logger.Debug("Parsing size parameter", "size_str", sizeStr)
		parsedSize, err := strconv.Atoi(sizeStr)
		if err != nil || parsedSize <= 0 || parsedSize > maxSize {
			h.logger.Warn("Invalid size parameter",
				"size_str", sizeStr,
				"error", err,
				"min", 1,
				"max", maxSize,
				"remote_addr", r.RemoteAddr,
			)
			http.Error(w, "Invalid size parameter: must be between 1 and 2048", http.StatusBadRequest)
			return
		}
		size = parsedSize
		h.logger.Debug("Size parameter parsed", "size", size)
	} else {
		h.logger.Debug("Using default size", "size", defaultSize)
	}

	h.logger.Debug("Calling QR generation service",
		"data_length", len(body),
		"size", size,
	)

	png, err := h.svc.Generate(body, size)
	if err != nil {
		h.logger.Error("failed to generate QR code",
			"error", err,
			"data_length", len(body),
			"size", size,
			"remote_addr", r.RemoteAddr,
		)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("QR code generated successfully",
		"png_size", len(png),
		"remote_addr", r.RemoteAddr,
	)

	w.Header().Set("Content-Type", "image/png")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		h.logger.Error("failed to write response",
			"error", err,
			"png_size", len(png),
			"remote_addr", r.RemoteAddr,
		)
		return
	}

	h.logger.Info("QR code request completed successfully",
		"data_length", len(body),
		"size", size,
		"output_size", len(png),
		"remote_addr", r.RemoteAddr,
	)
}

// HealthCheck handles GET/POST /health requests for liveness/readiness probes.
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Health check request received",
		"method", r.Method,
		"remote_addr", r.RemoteAddr,
	)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		h.logger.Error("failed to encode health check response",
			"error", err,
			"remote_addr", r.RemoteAddr,
		)
	}
}
