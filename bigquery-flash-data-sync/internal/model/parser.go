package model

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf8"

	"go.uber.org/zap"
)

// ParseDynamicRow scans a sql.Rows result into a DynamicRow structure.
// It handles various SQL types and converts them appropriately for BigQuery.
func ParseDynamicRow(rows *sql.Rows, logger *zap.Logger, dateFormat string) (*DynamicRow, error) {
	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("failed to get columns: %w", err)
	}

	values := make([]any, len(columns))
	valuePtrs := make([]any, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	if err := rows.Scan(valuePtrs...); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	for i, val := range values {
		values[i] = convertValue(val, dateFormat, logger)
	}

	return &DynamicRow{
		ColumnNames: columns,
		Values:      values,
	}, nil
}

func sanitizeInvalidUTF8(logger *zap.Logger, s string, originalLen int) string {
	logger.Debug("Invalid UTF-8 detected, sanitizing",
		zap.Int("original_length", originalLen),
	)
	return strings.ToValidUTF8(s, "")
}

// convertValue converts SQL values to appropriate Go types for BigQuery.
// It sanitizes invalid UTF-8 sequences and handles Unsigned Integer overflow.
func convertValue(val any, dateFormat string, logger *zap.Logger) any {
	if val == nil {
		return nil
	}

	switch v := val.(type) {
	case []byte:
		if !utf8.Valid(v) {
			return sanitizeInvalidUTF8(logger, string(v), len(v))
		}
		return string(v)
	case time.Time:
		return formatTimeForBigQuery(v, dateFormat)
	case int64, int32, int16, int8, int:
		return v
	case uint64:
		return safeUintToBigQuery(v, logger)
	case uint:
		return safeUintToBigQuery(uint64(v), logger)
	case uint32:
		return int64(v)
	case uint16:
		return int64(v)
	case uint8:
		return int64(v)
	case float64, float32:
		return v
	case bool:
		return v
	case string:
		if !utf8.ValidString(v) {
			return sanitizeInvalidUTF8(logger, v, len(v))
		}
		return v
	default:
		logger.Debug("Converting unknown type to string",
			zap.String("type", fmt.Sprintf("%T", v)))
		return fmt.Sprintf("%v", v)
	}
}

// formatTimeForBigQuery handles time serialization logic.
// It detects "date-only" values (midnight) and formats them strictly if a dateFormat is provided.
// Otherwise, it returns an RFC3339Nano timestamp.
func formatTimeForBigQuery(v time.Time, dateFormat string) any {
	if v.IsZero() {
		return nil
	}
	if v.Hour() == 0 && v.Minute() == 0 && v.Second() == 0 && v.Nanosecond() == 0 && dateFormat != "" {
		return v.Format(dateFormat)
	}
	return v.UTC().Format(time.RFC3339Nano)
}

// safeUintToBigQuery checks if a uint64 fits within BigQuery's INT64 (signed).
// If it fits, it returns int64. If it overflows, it returns string to preserve the value.
func safeUintToBigQuery(v uint64, logger *zap.Logger) any {
	if v > math.MaxInt64 {
		logger.Warn("uint value exceeds BigQuery INT64 max, converting to string",
			zap.Uint64("value", v))
		return fmt.Sprintf("%d", v)
	}
	return int64(v)
}
