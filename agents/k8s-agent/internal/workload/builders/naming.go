package builders

import (
	"fmt"
	"strings"
	"time"
)

// SanitizeName applies DNS-1123 conventions when constructing child resource names.
func SanitizeName(in string) string {
	if in == "" {
		return fmt.Sprintf("aegis-%d", time.Now().Unix())
	}
	cleaned := strings.ToLower(in)
	cleaned = strings.ReplaceAll(cleaned, "_", "-")
	cleaned = strings.ReplaceAll(cleaned, ".", "-")
	if len(cleaned) > 63 {
		cleaned = cleaned[:63]
	}
	return cleaned
}
