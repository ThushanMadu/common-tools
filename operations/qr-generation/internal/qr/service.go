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

// Package qr provides QR code generation functionality.
package qr

import (
	"fmt"
	"log/slog"

	"github.com/skip2/go-qrcode"
)

type Service interface {
	Generate(data []byte, size int) ([]byte, error)
}

type service struct {
	logger *slog.Logger
}

// NewService creates a new QR code generation service instance.
func NewService(logger *slog.Logger) Service {
	return &service{
		logger: logger,
	}
}

// Generate creates a QR code PNG image from the provided data with Medium error recovery (15%).
func (s *service) Generate(data []byte, size int) ([]byte, error) {
	s.logger.Debug("Starting QR code generation",
		"data_length", len(data),
		"size", size,
	)

	if len(data) == 0 {
		s.logger.Warn("QR code generation failed: empty data provided")
		return nil, fmt.Errorf("data cannot be empty")
	}

	if size <= 0 || size > 2048 {
		s.logger.Warn("QR code generation failed: invalid size",
			"size", size,
			"min", 1,
			"max", 2048,
		)
		return nil, fmt.Errorf("invalid size: must be between 1 and 2048")
	}

	s.logger.Debug("Encoding QR code",
		"recovery_level", "Medium",
		"data_preview", truncateString(string(data), 50),
	)

	png, err := qrcode.Encode(string(data), qrcode.Medium, size)
	if err != nil {
		s.logger.Error("Failed to encode QR code",
			"error", err,
			"data_length", len(data),
			"size", size,
		)
		return nil, fmt.Errorf("failed to encode QR code: %w", err)
	}

	s.logger.Debug("QR code generated successfully",
		"output_size_bytes", len(png),
		"image_dimensions", fmt.Sprintf("%dx%d", size, size),
	)

	return png, nil
}

// truncateString truncates a string to maxLen for safe logging.
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
