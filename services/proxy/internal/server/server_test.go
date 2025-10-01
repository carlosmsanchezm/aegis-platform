package server

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestCheckSessionActivity(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := Config{
		PrivilegedSessionTimeout: 10 * time.Millisecond,
		StandardSessionTimeout:   20 * time.Millisecond,
	}

	srv := &ProxyServer{
		log:          logger,
		cfg:          cfg,
		sessionStore: &sync.Map{},
	}

	jti := "test-session-123"

	err := srv.checkSessionActivity(jti, false)
	assert.NoError(t, err)

	time.Sleep(25 * time.Millisecond)
	err = srv.checkSessionActivity(jti, false)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "session timed out")

	privilegedJTI := "privileged-session-456"
	err = srv.checkSessionActivity(privilegedJTI, true)
	assert.NoError(t, err)

	time.Sleep(15 * time.Millisecond)
	err = srv.checkSessionActivity(privilegedJTI, true)
	assert.Error(t, err)
}
