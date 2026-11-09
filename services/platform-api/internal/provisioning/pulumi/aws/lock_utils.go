package aws

import (
	"errors"
	"os"
	"regexp"
	"strings"
	"time"
)

type s3LockInfo struct {
	Bucket    string
	Key       string
	CreatedBy string
	CreatedAt time.Time
}

var (
	s3PathRe        = regexp.MustCompile(`s3://([^/]+)/([^\s:]+)`)
	createdByRe     = regexp.MustCompile(`created by ([^\s]+)`)
	createdAtTimeRe = regexp.MustCompile(` at ([0-9]{4}-[0-9]{2}-[0-9]{2}T[^\s]+)`)
)

func parseLockInfoFromError(errMsg string) (*s3LockInfo, error) {
	msg := strings.TrimSpace(errMsg)
	if msg == "" {
		return nil, errors.New("empty error message")
	}

	m := s3PathRe.FindStringSubmatch(msg)
	if len(m) != 3 {
		return nil, errors.New("no S3 lock path found in error message")
	}

	info := &s3LockInfo{
		Bucket: strings.TrimSpace(m[1]),
		Key:    strings.TrimSpace(m[2]),
	}

	if cm := createdByRe.FindStringSubmatch(msg); len(cm) == 2 {
		info.CreatedBy = strings.TrimSpace(cm[1])
	}
	if tm := createdAtTimeRe.FindStringSubmatch(msg); len(tm) == 2 {
		if ts, err := time.Parse(time.RFC3339, strings.TrimSpace(tm[1])); err == nil {
			info.CreatedAt = ts
		}
	}

	return info, nil
}

func getLockStaleThreshold() time.Duration {
	const defaultThreshold = 30 * time.Minute

	raw := strings.TrimSpace(os.Getenv("AEGIS_PULUMI_LOCK_STALE_MINUTES"))
	if raw == "" {
		return defaultThreshold
	}

	minutes, err := time.ParseDuration(raw + "m")
	if err != nil {
		return defaultThreshold
	}
	if minutes <= 0 {
		return defaultThreshold
	}
	return minutes
}
