package workerpool

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VanceMichael/go-base-airbridge/internal/domain"
)

func TestStopCancelsRunningJobsBeforeWaiting(t *testing.T) {
	pool := New(1, 1)
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	pool.Start(parent)

	started := make(chan struct{})
	exited := make(chan error, 1)
	if err := pool.Submit(context.Background(), func(ctx context.Context) error {
		close(started)
		<-ctx.Done()
		exited <- ctx.Err()
		return ctx.Err()
	}); err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("job did not start")
	}

	stopped := make(chan struct{})
	go func() {
		pool.Stop()
		close(stopped)
	}()
	select {
	case <-stopped:
		select {
		case err := <-exited:
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("running job exit error = %v", err)
			}
		default:
			t.Fatal("Stop returned before the running job observed cancellation")
		}
	case <-time.After(150 * time.Millisecond):
		cancelParent()
		select {
		case <-stopped:
		case <-time.After(time.Second):
			t.Fatal("Stop remained blocked after test cleanup")
		}
		t.Fatal("Stop waited on a job whose context it did not cancel")
	}

	if err := pool.Submit(context.Background(), func(context.Context) error { return nil }); !errors.Is(err, domain.ErrState) {
		t.Fatalf("Submit() after stop error = %v", err)
	}
}
