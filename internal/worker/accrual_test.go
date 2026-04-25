package worker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type stubAccrualService struct {
	processPendingOrdersFunc func(ctx context.Context) error
}

func (s *stubAccrualService) ProcessPendingOrders(ctx context.Context) error {
	if s.processPendingOrdersFunc == nil {
		panic("unexpected call to stubAccrualService.ProcessPendingOrders")
	}

	return s.processPendingOrdersFunc(ctx)
}

func TestNewAccrualWorker_NilLogger(t *testing.T) {
	t.Parallel()

	accrualService := &stubAccrualService{
		processPendingOrdersFunc: func(ctx context.Context) error {
			return nil
		},
	}

	worker := NewAccrualWorker(accrualService, nil)

	require.NotNil(t, worker)
	require.NotNil(t, worker.logger)
	require.Same(t, accrualService, worker.accrualService)
}

func TestAccrualWorker_Run_ProcessPendingOrdersOnceAndStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	calls := 0
	accrualService := &stubAccrualService{
		processPendingOrdersFunc: func(ctx context.Context) error {
			calls++
			cancel()

			return nil
		},
	}

	worker := NewAccrualWorker(accrualService, nil)

	worker.Run(ctx)

	require.Equal(t, 1, calls)
}

func TestAccrualWorker_RunOnce_ContextCanceled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	called := false
	accrualService := &stubAccrualService{
		processPendingOrdersFunc: func(ctx context.Context) error {
			called = true

			return nil
		},
	}

	worker := NewAccrualWorker(accrualService, nil)

	worker.runOnce(ctx)

	require.False(t, called)
}

func TestAccrualWorker_RunOnce_ServiceError(t *testing.T) {
	t.Parallel()

	expectedErr := errors.New("process pending orders failed")

	called := false
	accrualService := &stubAccrualService{
		processPendingOrdersFunc: func(ctx context.Context) error {
			called = true

			return expectedErr
		},
	}

	worker := NewAccrualWorker(accrualService, nil)

	worker.runOnce(context.Background())

	require.True(t, called)
}
