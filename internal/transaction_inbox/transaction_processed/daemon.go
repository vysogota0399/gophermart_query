package transaction_processed

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type Daemon struct {
	lg           *logging.ZapLogger
	pollInterval time.Duration
	workersCount int64
	cfg          *transaction_inbox.Config

	cancaller    context.CancelFunc
	globalCtx    context.Context
	events       InboxEventsRepository
	orders       OrdersRepository
	transactions TransactionsRepository
}

type OrdersRepository interface {
	UpdateAccrualTX(ctx context.Context, in *models.Order, tx pgx.Tx) error
}
type InboxEventsRepository interface {
	ReserveTransactionProcessedEvent(ctx context.Context) (*models.TransactionEvent, error)
	SetState(ctx context.Context, uuid string, newState string) error
}

type TransactionsRepository interface {
	BeginTX(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	CommitTX(ctx context.Context, tx pgx.Tx) error
	RollbackTX(ctx context.Context, tx pgx.Tx) error
	Create(ctx context.Context, in *models.Transaction, tx pgx.Tx) error
}

func NewDaemon(
	lc fx.Lifecycle,
	events InboxEventsRepository,
	lg *logging.ZapLogger,
	cfg *transaction_inbox.Config,
	transactions TransactionsRepository,
	orders OrdersRepository,
) *Daemon {
	dmn := &Daemon{
		lg:           lg,
		pollInterval: time.Duration(cfg.OrderCreatedPollInterval) * time.Millisecond,
		events:       events,
		workersCount: cfg.WorkersCount,
		cfg:          cfg,
		transactions: transactions,
		orders:       orders,
	}
	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				dmn.Start()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				dmn.cancaller()
				return nil
			},
		},
	)

	return dmn
}

func (dmn *Daemon) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	dmn.cancaller = cancel
	dmn.globalCtx = dmn.lg.WithContextFields(ctx, zap.String("name", "transaction_created_events_daemon"))

	dmn.lg.DebugCtx(
		ctx,
		fmt.Sprintf("start processong %s events", repositories.OrderUpdatedEventName),
		zap.Any("config", dmn.cfg),
	)

	for i := 0; i < int(dmn.workersCount); i++ {
		wctx := dmn.lg.WithContextFields(ctx, zap.Int("worker_id", i))
		go func() {
			ticker := time.NewTicker(dmn.pollInterval)

			for {
				select {
				case <-wctx.Done():
					dmn.lg.DebugCtx(wctx, "daemon worker graceful shutdown")
					return
				case <-ticker.C:
					if err := dmn.processEvent(wctx); err != nil {
						dmn.lg.ErrorCtx(wctx, "process event finished error", zap.Error(err))
					}
				}
			}
		}()
	}
}

func (dmn *Daemon) processEvent(ctx context.Context) error {
	e, err := dmn.events.ReserveTransactionProcessedEvent(ctx)
	if err != nil {
		return fmt.Errorf("find first unprocessed event error %w", err)
	}

	if e == nil {
		return nil
	}

	ctx = dmn.lg.WithContextFields(ctx, zap.String("event_uuid", e.UUID))

	tx, err := dmn.transactions.BeginTX(ctx, pgx.TxOptions{})
	if err != nil {
		if err := dmn.events.SetState(ctx, e.UUID, models.TransactionEventFailedState); err != nil {
			return fmt.Errorf("set processing event  %s state error %w", models.TransactionEventFailedState, err)
		}

		return fmt.Errorf("find first unprocessed event error %w", err)
	}
	defer dmn.transactions.RollbackTX(ctx, tx)

	transaction := &models.Transaction{
		UUID:        e.Meta.UUID,
		OrderNumber: e.Meta.OrderNumber,
		AccountID:   e.Meta.AccountID,
		Amount:      e.Meta.Amount,
		Operation:   e.Meta.Operation,
		ProcessedAt: e.Meta.ProcessedAt,
	}
	if err := dmn.transactions.Create(ctx, transaction, tx); err != nil {
		if err := dmn.events.SetState(ctx, e.UUID, models.TransactionEventFailedState); err != nil {
			return fmt.Errorf("set processing event %s state error %w", models.TransactionEventFailedState, err)
		}

		return fmt.Errorf("create transaction(%+v) error %w", transaction, err)
	}

	order := &models.Order{
		Number:  e.Meta.OrderNumber,
		Accrual: e.Meta.Amount,
	}

	if err := dmn.orders.UpdateAccrualTX(ctx, order, tx); err != nil {
		if err := dmn.events.SetState(ctx, e.UUID, models.TransactionEventFailedState); err != nil {
			return fmt.Errorf("set processing event %s state error %w", models.TransactionEventFailedState, err)
		}

		return fmt.Errorf("update accrual failed error %w", err)
	}

	if err := dmn.events.SetState(ctx, e.UUID, models.TransactionEventFinishedState); err != nil {
		return fmt.Errorf("set processing event %s state error %w", models.TransactionEventFinishedState, err)
	}

	return dmn.transactions.CommitTX(ctx, tx)
}
