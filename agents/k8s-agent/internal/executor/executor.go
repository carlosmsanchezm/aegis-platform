package executor

import (
	"context"

	aegis "github.com/yourorg/aegis/proto/aegis/v1"
)

type Result struct {
	Status  string
	URL     string
	Backend string
	Err     error
}

type Executor interface {
	RunWorkspace(ctx context.Context, w *aegis.Workload) Result
	RunTraining(ctx context.Context, w *aegis.Workload) Result
}
