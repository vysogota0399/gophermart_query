package order_created

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
	SaveOrderCreated(ctx context.Context, in *models.OrderCreatedEvent) error
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
	lg.DebugCtx(context.Background(), "start order created events consumer", zap.String("consumer_group", cfg.KafkaOrderCreatedGroupID), zap.Any("config", cfg))

	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:                cfg.KafkaOrderCreatedGroupID,
		PartitionWatchInterval: time.Duration(cfg.KafkaOrderCreatedPartitionWatchInterval) * time.Millisecond,
		Brokers:                globalCFG.KafkaBrokers,
		Topic:                  cfg.KafkaOrderCreatedTopic,
		MinBytes:               10e2, // 1KB
		MaxBytes:               10e6, // 10MB
		MaxWait:                time.Duration(cfg.KafkaOrderCreatedMaxWaitInterval) * time.Millisecond,
		ErrorLogger:            errLogger,
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
				cns.lg.ErrorCtx(ctx, "order_created/consumer: fetch message error", zap.Error(err))
			}
		}
	}
}

func (cns *Consumer) processMessage(ctx context.Context) error {
	m, err := cns.reader.FetchMessage(cns.globalCtx)
	if err != nil {
		return fmt.Errorf("order_created/consumer: fetch message error %w", err)
	}

	payload := events.OrderCreated{}

	if err := proto.Unmarshal(m.Value, &payload); err != nil {
		return fmt.Errorf("order_created/consumer: unmarshal message error %w", err)
	}

	cns.lg.InfoCtx(ctx, "consumed message", zap.Any("message", &payload))

	updatedAt, err := time.Parse(time.RFC3339Nano, payload.UploadedAt)
	if err != nil {
		return fmt.Errorf("order_created/consumer: unmarshal message error %w", err)
	}

	if err := cns.events.SaveOrderCreated(
		ctx,
		&models.OrderCreatedEvent{
			UUID:  payload.EventUuid,
			State: models.OrderEventNewState,
			Name:  repositories.OrderCreatedEventName,
			Meta: &models.OrderCreatedEventMeta{
				UUID:       payload.Uuid,
				Number:     payload.Number,
				State:      payload.State,
				UploadedAt: updatedAt,
				AccountID:  payload.AccountId,
			},
		},
	); err != nil {
		return fmt.Errorf("order_created/consumer: save message error %w", err)
	}

	if err := cns.reader.CommitMessages(ctx, m); err != nil {
		return fmt.Errorf("order_created/consumer: failed to commit messages %w", err)
	}

	return nil
}
