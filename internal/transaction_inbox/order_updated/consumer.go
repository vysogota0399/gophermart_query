package order_updated

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
	"github.com/vysogota0399/gophermart_protos/gen/events"
	"github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type Consumer struct {
	lg        *logging.ZapLogger
	reader    *kafka.Reader
	events    ConsumerInboxEventsRepository
	cancaller context.CancelFunc
	globalCtx context.Context
}

type ConsumerInboxEventsRepository interface {
	SaveOrderUpdated(ctx context.Context, in *models.OrderUpdatedEvent) error
}

func NewConsumer(
	lc fx.Lifecycle,
	lg *logging.ZapLogger,
	cfg *transaction_inbox.Config,
	globalCFG *config.Config,
	errLogger *logging.KafkaErrorLogger,
	logger *logging.KafkaLogger,
	events ConsumerInboxEventsRepository,
) *Consumer {
	lg.DebugCtx(context.Background(), "start order updated events consumer", zap.String("consumer_group", cfg.KafkaOrderUpdatedGroupID), zap.Any("config", cfg))

	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:                cfg.KafkaOrderUpdatedGroupID,
		PartitionWatchInterval: time.Duration(cfg.KafkaOrderUpdatedPartitionWatchInterval) * time.Millisecond,
		Brokers:                globalCFG.KafkaBrokers,
		Topic:                  cfg.KafkaOrderUpdatedTopic,
		MinBytes:               10e2, // 1KB
		MaxBytes:               10e6, // 10MB
		ErrorLogger:            errLogger,
		MaxWait:                time.Duration(cfg.KafkaOrderUpdatedMaxWaitInterval) * time.Millisecond,
		Logger:                 logger,
	})

	cns := &Consumer{
		lg:     lg,
		reader: r,
		events: events,
	}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				for {
					go cns.consume()
					return nil
				}
			},
			OnStop: func(ctx context.Context) error {
				return cns.reader.Close()
			},
		},
	)

	return cns
}

func (cns *Consumer) consume() {
	ctx, cancel := context.WithCancel(context.Background())
	cns.globalCtx = ctx
	cns.cancaller = cancel

	for {
		select {
		case <-ctx.Done():
			cns.lg.DebugCtx(ctx, "consumer graceful shutdown")
			return
		default:
			if err := cns.processMessage(cns.globalCtx); err != nil {
				cns.lg.ErrorCtx(ctx, "order_updated/consumer: fetch message error", zap.Error(err))
			}
		}
	}
}

func (cns *Consumer) processMessage(ctx context.Context) error {
	m, err := cns.reader.FetchMessage(cns.globalCtx)
	if err != nil {
		return fmt.Errorf("order_updated/consumer: fetch message error %w", err)
	}

	payload := events.OrderUpdated{}

	if err := proto.Unmarshal(m.Value, &payload); err != nil {
		return fmt.Errorf("order_updated/consumer: unmarshal message error %w", err)
	}

	cns.lg.InfoCtx(ctx, "consumed message", zap.Any("message", &payload))

	if err := cns.events.SaveOrderUpdated(
		ctx,
		&models.OrderUpdatedEvent{
			UUID:  payload.EventUuid,
			State: models.OrderEventNewState,
			Name:  repositories.OrderUpdatedEventName,
			Meta: &models.OrderUpdatedEventMeta{
				UUID:  payload.Uuid,
				State: payload.State,
			},
		},
	); err != nil {
		return fmt.Errorf("order_updated/consumer: save message error %w", err)
	}

	if err := cns.reader.CommitMessages(ctx, m); err != nil {
		return fmt.Errorf("order_updated/consumer: failed to commit messages %w", err)
	}

	return nil
}
