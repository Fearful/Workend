package run

import (
	"regexp"
	"strings"
)

var sensitivePatterns = []*regexp.Regexp{
	regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(secret|token)\s*[=:]\s*\S+`),
	regexp.MustCompile(`(?i)(aws_secret_access_key|aws_access_key_id)\s*[=:]\s*\S+`),
	regexp.MustCompile(`ghp_[A-Za-z0-9]{36,}`),
	regexp.MustCompile(`gho_[A-Za-z0-9]{36,}`),
	regexp.MustCompile(`glpat-[A-Za-z0-9\-]{20,}`),
	regexp.MustCompile(`sk-[A-Za-z0-9]{20,}`),
	regexp.MustCompile(`(?i)bearer\s+[A-Za-z0-9\-._~+/]+=*`),
}

// MaskSecrets replaces sensitive values (tokens, passwords, API keys) in
// the input string with asterisks, preserving only the first 4 characters
// of longer matches for debugging context.
func MaskSecrets(input string) string {
	result := input
	for _, pat := range sensitivePatterns {
		result = pat.ReplaceAllStringFunc(result, func(match string) string {
			if len(match) <= 8 {
				return strings.Repeat("*", len(match))
			}
			return match[:4] + strings.Repeat("*", len(match)-4)
		})
	}
	return result
}
