package worker

import (
	"context"
	"time"

	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

type AccrualWorker struct {
	accrualService service.AccrualService
	logger         *zap.Logger
}

func NewAccrualWorker(
	accrualService service.AccrualService,
	logger *zap.Logger,
) *AccrualWorker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &AccrualWorker{
		accrualService: accrualService,
		logger:         logger,
	}
}

func (w *AccrualWorker) Run(ctx context.Context) {
	w.logger.Info(" --> accrual worker started")

	w.runOnce(ctx)

	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("accrual worker stopped <--")
			return

		case <-ticker.C:
			w.runOnce(ctx)
		}
	}
}

func (w *AccrualWorker) runOnce(ctx context.Context) {
	if err := ctx.Err(); err != nil {
		return
	}

	err := w.accrualService.ProcessPendingOrders(ctx)
	if err != nil {
		w.logger.Warn(" *** accrual worker run once failed", zap.Error(err))
	}
}
