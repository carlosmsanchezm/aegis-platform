package aws

import (
	"testing"
	"time"
)

func TestParseLockInfoFromError(t *testing.T) {
	tests := []struct {
		name        string
		errMsg      string
		wantBucket  string
		wantKey     string
		wantCreator string
		wantErr     bool
	}{
		{
			name:        "valid S3 lock error",
			errMsg:      `the stack is currently locked by 1 lock(s). Either wait for the other process(es) to end or delete the lock file with "pulumi cancel". s3://aegis-pulumi-state-dev/aegis/pulumi/.pulumi/locks/organization/aegis-platform-db-1/aegis-db-1-us-east-1/d3adf854-aca8-456e-9e8c-e1b25bf3c4f6.json: created by carlossanchez@Carloss-Mac-mini.local (pid 28296) at 2025-12-12T15:11:49-05:00`,
			wantBucket:  "aegis-pulumi-state-dev",
			wantKey:     "aegis/pulumi/.pulumi/locks/organization/aegis-platform-db-1/aegis-db-1-us-east-1/d3adf854-aca8-456e-9e8c-e1b25bf3c4f6.json",
			wantCreator: "carlossanchez@Carloss-Mac-mini.local",
			wantErr:     false,
		},
		{
			name:    "no S3 path in error",
			errMsg:  "some random error without s3 path",
			wantErr: true,
		},
		{
			name:       "S3 path without timestamp",
			errMsg:     `stack locked by s3://mybucket/path/to/lock.json: created by user@host`,
			wantBucket: "mybucket",
			wantKey:    "path/to/lock.json",
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := parseLockInfoFromError(tt.errMsg)
			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if info.Bucket != tt.wantBucket {
				t.Errorf("bucket = %q, want %q", info.Bucket, tt.wantBucket)
			}
			if info.Key != tt.wantKey {
				t.Errorf("key = %q, want %q", info.Key, tt.wantKey)
			}
			if tt.wantCreator != "" && info.CreatedBy != tt.wantCreator {
				t.Errorf("createdBy = %q, want %q", info.CreatedBy, tt.wantCreator)
			}
		})
	}
}

func TestGetLockStaleThreshold(t *testing.T) {
	// Default should be 30 minutes
	threshold := getLockStaleThreshold()
	if threshold != 30*time.Minute {
		t.Errorf("default threshold = %v, want 30m", threshold)
	}

	// Setting env var should change it
	t.Setenv("AEGIS_PULUMI_LOCK_STALE_MINUTES", "10")
	threshold = getLockStaleThreshold()
	if threshold != 10*time.Minute {
		t.Errorf("threshold with env = %v, want 10m", threshold)
	}

	// Invalid env var should fallback to default
	t.Setenv("AEGIS_PULUMI_LOCK_STALE_MINUTES", "invalid")
	threshold = getLockStaleThreshold()
	if threshold != 30*time.Minute {
		t.Errorf("threshold with invalid env = %v, want 30m", threshold)
	}
}

func TestIsLockError(t *testing.T) {
	tests := []struct {
		name   string
		errMsg string
		isLock bool
	}{
		{
			name:   "stack currently locked",
			errMsg: "error: could not create stack: the stack is currently locked by 1 lock(s)",
			isLock: true,
		},
		{
			name:   "currently locked by",
			errMsg: "stack currently locked by another process",
			isLock: true,
		},
		{
			name:   "lock(s) keyword",
			errMsg: "detected 2 lock(s) on the stack",
			isLock: true,
		},
		{
			name:   "not a lock error",
			errMsg: "connection timeout",
			isLock: false,
		},
		{
			name:   "nil error message",
			errMsg: "",
			isLock: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.errMsg != "" {
				err = &testError{msg: tt.errMsg}
			}
			got := isLockError(err)
			if got != tt.isLock {
				t.Errorf("isLockError(%q) = %v, want %v", tt.errMsg, got, tt.isLock)
			}
		})
	}
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}

