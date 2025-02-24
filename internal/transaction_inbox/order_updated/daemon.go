package order_updated

import (
	"context"
	"fmt"
	"time"

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

	cancaller context.CancelFunc
	globalCtx context.Context
	events    InboxEventsRepository
	orders    OrdersRepository
}

type InboxEventsRepository interface {
	ReserveOrderUpdatedEvent(ctx context.Context) (*models.OrderUpdatedEvent, error)
	SetState(ctx context.Context, uuid string, newState string) error
}

type OrdersRepository interface {
	UpdateState(ctx context.Context, in *models.Order) error
}

func NewDaemon(
	lc fx.Lifecycle,
	events InboxEventsRepository,
	lg *logging.ZapLogger,
	cfg *transaction_inbox.Config,
	orders OrdersRepository,
) *Daemon {
	dmn := &Daemon{
		lg:           lg,
		pollInterval: time.Duration(cfg.OrderCreatedPollInterval) * time.Millisecond,
		events:       events,
		workersCount: cfg.WorkersCount,
		cfg:          cfg,
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
	dmn.globalCtx = dmn.lg.WithContextFields(ctx, zap.String("name", "order_updated_events_daemon"))

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
	e, err := dmn.events.ReserveOrderUpdatedEvent(ctx)
	if err != nil {
		return fmt.Errorf("find first unprocessed event error %w", err)
	}

	if e == nil {
		return nil
	}

	ctx = dmn.lg.WithContextFields(ctx, zap.String("event_uuid", e.UUID))

	if err := dmn.orders.UpdateState(
		ctx,
		&models.Order{
			UUID:  e.Meta.UUID,
			State: e.Meta.State,
		},
	); err != nil {
		if err := dmn.events.SetState(ctx, e.UUID, models.OrderEventFailedState); err != nil {
			return fmt.Errorf("set processing event state error %w", err)
		}
	}

	if err := dmn.events.SetState(ctx, e.UUID, models.OrderEventFinishedState); err != nil {
		return fmt.Errorf("set processing event state error %w", err)
	}

	return nil
}
