package common

import (
	"fmt"
	"math"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// DetectDataType analyzes a slice of values and determines the column type
func DetectDataType(values []string) DataType {
	if len(values) == 0 {
		return TypeEmpty
	}

	var (
		stringCount  int
		numberCount  int
		dateCount    int
		booleanCount int
		emptyCount   int
	)

	for _, val := range values {
		trimmed := strings.TrimSpace(val)

		// Check for empty
		if trimmed == "" {
			emptyCount++
			continue
		}

		// Check for boolean
		lower := strings.ToLower(trimmed)
		if lower == "true" || lower == "false" || lower == "yes" || lower == "no" || lower == "1" || lower == "0" {
			booleanCount++
			continue
		}

		// Check for number
		if _, err := strconv.ParseFloat(trimmed, 64); err == nil {
			numberCount++
			continue
		}

		// Check for date (various formats)
		if IsDateValue(trimmed) {
			dateCount++
			continue
		}

		// Default to string
		stringCount++
	}

	total := len(values) - emptyCount
	if total == 0 {
		return TypeEmpty
	}

	// Determine primary type (>80% threshold)
	threshold := float64(total) * 0.8

	if float64(numberCount) >= threshold {
		return TypeNumber
	}
	if float64(dateCount) >= threshold {
		return TypeDate
	}
	if float64(booleanCount) >= threshold {
		return TypeBoolean
	}
	if float64(stringCount) >= threshold {
		return TypeString
	}

	return TypeMixed
}

// IsDateValue checks if a string looks like a date
func IsDateValue(val string) bool {
	// Common date formats to try
	formats := []string{
		"2006-01-02",
		"2006/01/02",
		"01/02/2006",
		"02/01/2006",
		"Jan 2, 2006",
		"2 Jan 2006",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04:05",
		"01-02-2006",
		"02-01-2006",
	}

	trimmed := strings.TrimSpace(val)
	for _, format := range formats {
		if _, err := time.Parse(format, trimmed); err == nil {
			return true
		}
	}

	// Check for ISO 8601 format
	iso8601 := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}(T\d{2}:\d{2}:\d{2})?`)
	return iso8601.MatchString(trimmed)
}

// TruncateString truncates a string to a maximum length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return "..."
	}
	return s[:maxLen-3] + "..."
}

// GetUniqueValues returns unique values from a slice
func GetUniqueValues(values []string) []string {
	seen := make(map[string]bool)
	var unique []string

	for _, val := range values {
		if !seen[val] {
			seen[val] = true
			unique = append(unique, val)
		}
	}

	return unique
}

// CountNulls counts empty or null values
func CountNulls(values []string) int {
	count := 0
	for _, val := range values {
		trimmed := strings.TrimSpace(val)
		if trimmed == "" || strings.ToLower(trimmed) == "null" || strings.ToLower(trimmed) == "nil" {
			count++
		}
	}
	return count
}

// FormatTable creates an ASCII table for display
func FormatTable(headers []string, rows [][]string, maxWidth int) string {
	if len(headers) == 0 || len(rows) == 0 {
		return ""
	}

	// Calculate column widths
	colWidths := make([]int, len(headers))
	for i, header := range headers {
		colWidths[i] = len(header)
	}

	for _, row := range rows {
		for i, cell := range row {
			if i < len(colWidths) && len(cell) > colWidths[i] {
				colWidths[i] = len(cell)
			}
		}
	}

	// Apply max width per column
	maxColWidth := maxWidth / len(headers)
	for i := range colWidths {
		if colWidths[i] > maxColWidth {
			colWidths[i] = maxColWidth
		}
	}

	var result strings.Builder

	// Top border
	result.WriteString("┌")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			result.WriteString("┬")
		}
	}
	result.WriteString("┐\n")

	// Headers
	result.WriteString("│")
	for i, header := range headers {
		result.WriteString(" ")
		result.WriteString(PadRight(TruncateString(header, colWidths[i]), colWidths[i]))
		result.WriteString(" │")
	}
	result.WriteString("\n")

	// Header separator
	result.WriteString("├")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			result.WriteString("┼")
		}
	}
	result.WriteString("┤\n")

	// Data rows
	for _, row := range rows {
		result.WriteString("│")
		for i := 0; i < len(headers); i++ {
			cell := ""
			if i < len(row) {
				cell = row[i]
			}
			result.WriteString(" ")
			result.WriteString(PadRight(TruncateString(cell, colWidths[i]), colWidths[i]))
			result.WriteString(" │")
		}
		result.WriteString("\n")
	}

	// Bottom border
	result.WriteString("└")
	for i, width := range colWidths {
		result.WriteString(strings.Repeat("─", width+2))
		if i < len(colWidths)-1 {
			result.WriteString("┴")
		}
	}
	result.WriteString("┘")

	return result.String()
}

// PadRight pads a string to the right with spaces
func PadRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// FormatPercentage formats a percentage nicely
func FormatPercentage(count, total int) string {
	if total == 0 {
		return "0%"
	}
	percentage := float64(count) * 100.0 / float64(total)
	return fmt.Sprintf("%.1f%%", percentage)
}

// GenerateRandomIndices generates n random indices from 0 to max-1
func GenerateRandomIndices(n, max int) []int {
	if n >= max {
		// Return all indices
		indices := make([]int, max)
		for i := range indices {
			indices[i] = i
		}
		return indices
	}

	// Initialize random with current time
	rand.Seed(time.Now().UnixNano())

	// Generate random permutation and take first n
	perm := rand.Perm(max)
	indices := perm[:n]

	// Sort indices for better display (optional)
	// You could remove this if you want truly random order
	// sort.Ints(indices)

	return indices
}

// Min returns the minimum of two integers
func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Max returns the maximum of two integers
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// Round rounds a float to n decimal places
func Round(val float64, precision int) float64 {
	ratio := math.Pow(10, float64(precision))
	return math.Round(val*ratio) / ratio
}