package worker

import (
	"context"
	"errors"
	"time"

	"github.com/xhrobj/gophermart/internal/service"
	"go.uber.org/zap"
)

// AccrualWorker периодически запускает обработку заказов через внешний сервис accrual.
type AccrualWorker struct {
	accrualService service.PendingOrdersProcessor
	logger         *zap.Logger
}

// NewAccrualWorker создаёт worker для периодического polling заказов в accrual.
func NewAccrualWorker(
	accrualService service.PendingOrdersProcessor,
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
	if err == nil {
		return
	}

	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}

	w.logger.Warn(" *** accrual worker run once failed", zap.Error(err))
}
