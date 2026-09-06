package codingtool

import (
	"context"
	"errors"
	"time"
)

// ErrExecutableDetectionTimedOut identifies executable detection that did not
// finish before the caller's deadline. It is safe to present as a temporary
// readiness state rather than a raw adapter error.
var ErrExecutableDetectionTimedOut = errors.New("coding-tool executable detection timed out")

// DetectExecutableDetailedWithin runs executable detection with a deadline.
// The result channel is buffered so a non-context-aware adapter can finish
// after the caller has moved on without blocking a goroutine on delivery.
// Adapters remain synchronous by contract; callers that need bounded UI work
// should use this helper rather than starting unmanaged detection goroutines.
func DetectExecutableDetailedWithin(ctx context.Context, timeout time.Duration, tool CodingTool, config Config) (ExecutableDetection, error) {
	if err := ctx.Err(); err != nil {
		return ExecutableDetection{}, err
	}
	if timeout <= 0 {
		return detectExecutableDetailed(tool, config)
	}

	type result struct {
		detection ExecutableDetection
		err       error
	}
	completed := make(chan result, 1)
	go func() {
		detection, err := detectExecutableDetailed(tool, config)
		completed <- result{detection: detection, err: err}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case result := <-completed:
		return result.detection, result.err
	case <-ctx.Done():
		return ExecutableDetection{}, ctx.Err()
	case <-timer.C:
		return ExecutableDetection{}, ErrExecutableDetectionTimedOut
	}
}

func detectExecutableDetailed(tool CodingTool, config Config) (ExecutableDetection, error) {
	if detailed, ok := tool.(DetailedExecutableDetector); ok {
		return detailed.DetectExecutableDetailed(config)
	}
	executable, err := tool.DetectExecutable(config)
	return ExecutableDetection{Executable: executable}, err
}
