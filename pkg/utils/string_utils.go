package utils

import (
	"strings"
	"unicode"
)

// Contains checks if a slice contains a specific string
func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ContainsIgnoreCase checks if a slice contains a string (case insensitive)
func ContainsIgnoreCase(slice []string, item string) bool {
	itemLower := strings.ToLower(item)
	for _, s := range slice {
		if strings.ToLower(s) == itemLower {
			return true
		}
	}
	return false
}

// TrimSpaceAll trims whitespace from all strings in a slice
func TrimSpaceAll(slice []string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		result[i] = strings.TrimSpace(s)
	}
	return result
}

// FilterEmpty removes empty strings from a slice
func FilterEmpty(slice []string) []string {
	result := make([]string, 0, len(slice))
	for _, s := range slice {
		if strings.TrimSpace(s) != "" {
			result = append(result, s)
		}
	}
	return result
}

// SplitAndTrim splits a string by delimiter and trims whitespace
func SplitAndTrim(s, delimiter string) []string {
	parts := strings.Split(s, delimiter)
	return FilterEmpty(TrimSpaceAll(parts))
}

// ToSnakeCase converts a string to snake_case
func ToSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// ToCamelCase converts a string to camelCase
func ToCamelCase(s string) string {
	words := strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})

	if len(words) == 0 {
		return s
	}

	var result strings.Builder
	result.WriteString(strings.ToLower(words[0]))

	for _, word := range words[1:] {
		if len(word) > 0 {
			result.WriteString(strings.ToUpper(string(word[0])))
			if len(word) > 1 {
				result.WriteString(strings.ToLower(word[1:]))
			}
		}
	}

	return result.String()
}

// ToPascalCase converts a string to PascalCase
func ToPascalCase(s string) string {
	camel := ToCamelCase(s)
	if len(camel) == 0 {
		return camel
	}
	return strings.ToUpper(string(camel[0])) + camel[1:]
}

// Truncate truncates a string to maxLength with optional suffix
func Truncate(s string, maxLength int, suffix string) string {
	if len(s) <= maxLength {
		return s
	}

	if len(suffix) >= maxLength {
		return s[:maxLength]
	}

	return s[:maxLength-len(suffix)] + suffix
}

// RemoveDuplicates removes duplicate strings from a slice while preserving order
func RemoveDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(slice))

	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}

	return result
}

// IsEmpty checks if a string is empty or contains only whitespace
func IsEmpty(s string) bool {
	return strings.TrimSpace(s) == ""
}

// SanitizeForFilename removes or replaces characters that are invalid in filenames
func SanitizeForFilename(s string) string {
	// Replace invalid characters with underscores
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := s

	for _, char := range invalid {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Trim spaces and dots from the ends
	result = strings.Trim(result, " .")

	return result
}

// ExtractWords extracts words from a string, splitting on non-alphanumeric characters
func ExtractWords(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
}

// JoinNonEmpty joins non-empty strings with a separator
func JoinNonEmpty(slice []string, separator string) string {
	nonEmpty := FilterEmpty(slice)
	return strings.Join(nonEmpty, separator)
}

// PadLeft pads a string to the left with a character to reach the specified length
func PadLeft(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}

	padding := strings.Repeat(string(pad), length-len(s))
	return padding + s
}

// PadRight pads a string to the right with a character to reach the specified length
func PadRight(s string, length int, pad rune) string {
	if len(s) >= length {
		return s
	}

	padding := strings.Repeat(string(pad), length-len(s))
	return s + padding
}

// ReverseString reverses a string
func ReverseString(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}
