package codingtool_test

import (
	"context"
	"errors"
	"testing"
	"time"

	codingtool "devctx/packages/core/codingtool"
)

func TestDetectExecutableDetailedWithinReturnsTimeout(t *testing.T) {
	release := make(chan struct{})
	defer close(release)

	_, err := codingtool.DetectExecutableDetailedWithin(context.Background(), time.Millisecond, delayedDetectionTool{release: release}, codingtool.Config{})
	if !errors.Is(err, codingtool.ErrExecutableDetectionTimedOut) {
		t.Fatalf("error = %v, want timeout", err)
	}
}

func TestDetectExecutableDetailedWithinReturnsCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := codingtool.DetectExecutableDetailedWithin(ctx, time.Second, delayedDetectionTool{}, codingtool.Config{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("error = %v, want cancellation", err)
	}
}

func TestDetectExecutableDetailedWithinPreservesDetailedResult(t *testing.T) {
	detection, err := codingtool.DetectExecutableDetailedWithin(context.Background(), time.Second, detailedDetectionTool{}, codingtool.Config{})
	if err != nil {
		t.Fatalf("detect executable: %v", err)
	}
	if detection.Executable != "/opt/code" || detection.Platform != "linux" || detection.Source != codingtool.ExecutableDetectionConfigured {
		t.Fatalf("detection = %#v", detection)
	}
}

type delayedDetectionTool struct{ release <-chan struct{} }

func (delayedDetectionTool) ID() codingtool.ID { return "delayed" }
func (t delayedDetectionTool) DetectExecutable(codingtool.Config) (codingtool.Executable, error) {
	if t.release != nil {
		<-t.release
	}
	return "/opt/code", nil
}
func (delayedDetectionTool) BuildLaunchCommand(codingtool.CommandRequest) (codingtool.Command, error) {
	return codingtool.Command{}, nil
}

type detailedDetectionTool struct{ delayedDetectionTool }

func (detailedDetectionTool) DetectExecutableDetailed(codingtool.Config) (codingtool.ExecutableDetection, error) {
	return codingtool.ExecutableDetection{Executable: "/opt/code", Platform: "linux", Source: codingtool.ExecutableDetectionConfigured}, nil
}
