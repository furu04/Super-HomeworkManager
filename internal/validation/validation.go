package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var MaxLengths = map[string]int{
	"title":       200,
	"description": 5000,
	"subject":     100,
	"priority":    20,
}

var xssPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)<\s*script`),
	regexp.MustCompile(`(?i)</\s*script`),
	regexp.MustCompile(`(?i)javascript\s*:`),
	regexp.MustCompile(`(?i)on\w+\s*=`),
	regexp.MustCompile(`(?i)<\s*iframe`),
	regexp.MustCompile(`(?i)<\s*object`),
	regexp.MustCompile(`(?i)<\s*embed`),
	regexp.MustCompile(`(?i)<\s*svg[^>]*on\w+\s*=`),
	regexp.MustCompile(`(?i)data\s*:\s*text/html`),
	regexp.MustCompile(`(?i)<\s*img[^>]*on\w+\s*=`),
	regexp.MustCompile(`(?i)expression\s*\(`),
	regexp.MustCompile(`(?i)alert\s*\(`),
	regexp.MustCompile(`(?i)confirm\s*\(`),
	regexp.MustCompile(`(?i)prompt\s*\(`),
	regexp.MustCompile(`(?i)document\s*\.\s*cookie`),
	regexp.MustCompile(`(?i)document\s*\.\s*location`),
	regexp.MustCompile(`(?i)window\s*\.\s*location`),
}

var sqlInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)'\s*or\s+`),
	regexp.MustCompile(`(?i)'\s*and\s+`),
	regexp.MustCompile(`(?i)"\s*or\s+`),
	regexp.MustCompile(`(?i)"\s*and\s+`),
	regexp.MustCompile(`(?i)union\s+(all\s+)?select`),
	regexp.MustCompile(`(?i);\s*(drop|delete|update|insert|alter|truncate)\s+`),
	regexp.MustCompile(`(?i)--\s*$`),
	regexp.MustCompile(`(?i)/\*.*\*/`),
	regexp.MustCompile(`(?i)'\s*;\s*`),
	regexp.MustCompile(`(?i)exec\s*\(`),
	regexp.MustCompile(`(?i)xp_\w+`),
	regexp.MustCompile(`(?i)load_file\s*\(`),
	regexp.MustCompile(`(?i)into\s+(out|dump)file`),
	regexp.MustCompile(`(?i)benchmark\s*\(`),
	regexp.MustCompile(`(?i)sleep\s*\(\s*\d`),
	regexp.MustCompile(`(?i)waitfor\s+delay`),
	regexp.MustCompile(`(?i)1\s*=\s*1`),
	regexp.MustCompile(`(?i)'1'\s*=\s*'1`),
}

var pathTraversalPatterns = []*regexp.Regexp{
	regexp.MustCompile(`\.\.[\\/]`),
	regexp.MustCompile(`\.\.%2[fF]`),
	regexp.MustCompile(`%2e%2e[\\/]`),
	regexp.MustCompile(`\.\./`),
	regexp.MustCompile(`\.\.\\`),
}

var commandInjectionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*;`),
	regexp.MustCompile(`;\s*\w+`),
	regexp.MustCompile(`\|\s*\w+`),
	regexp.MustCompile("`[^`]+`"),
	regexp.MustCompile(`\$\([^)]+\)`),
	regexp.MustCompile(`&&\s*\w+`),
	regexp.MustCompile(`\|\|\s*\w+`),
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func ValidateAssignmentInput(title, description, subject, priority string) error {
	if err := ValidateField("title", title, true); err != nil {
		return err
	}
	if err := ValidateField("description", description, false); err != nil {
		return err
	}
	if err := ValidateField("subject", subject, false); err != nil {
		return err
	}
	if err := ValidateField("priority", priority, false); err != nil {
		return err
	}
	return nil
}

func ValidateField(fieldName, value string, required bool) error {
	if required && strings.TrimSpace(value) == "" {
		return &ValidationError{Field: fieldName, Message: "必須項目です"}
	}

	if value == "" {
		return nil
	}

	if maxLen, ok := MaxLengths[fieldName]; ok {
		if len(value) > maxLen {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("最大%d文字までです", maxLen),
			}
		}
	}

	if fieldName != "description" {
		for _, r := range value {
			if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
				return &ValidationError{
					Field:   fieldName,
					Message: "不正な制御文字が含まれています",
				}
			}
		}
	}

	for _, pattern := range xssPatterns {
		if pattern.MatchString(value) {
			return &ValidationError{
				Field:   fieldName,
				Message: "潜在的に危険なHTMLタグまたはスクリプトが含まれています",
			}
		}
	}

	for _, pattern := range sqlInjectionPatterns {
		if pattern.MatchString(value) {
			return &ValidationError{
				Field:   fieldName,
				Message: "潜在的に危険なSQL構文が含まれています",
			}
		}
	}

	for _, pattern := range pathTraversalPatterns {
		if pattern.MatchString(value) {
			return &ValidationError{
				Field:   fieldName,
				Message: "不正なパス文字列が含まれています",
			}
		}
	}

	for _, pattern := range commandInjectionPatterns {
		if pattern.MatchString(value) {
			return &ValidationError{
				Field:   fieldName,
				Message: "潜在的に危険なコマンド構文が含まれています",
			}
		}
	}

	return nil
}

func SanitizeString(s string) string {
	s = strings.ReplaceAll(s, "\x00", "")
	s = strings.TrimSpace(s)
	return s
}
