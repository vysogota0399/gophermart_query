package transaction_processed

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
	SaveTransactionCreated(ctx context.Context, in *models.TransactionEvent) error
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
	lg.DebugCtx(context.Background(), "start transaction created events consumer", zap.String("consumer_group", cfg.KafkaTransactionProcessedGroupID), zap.Any("config", cfg))

	r := kafka.NewReader(kafka.ReaderConfig{
		GroupID:                cfg.KafkaTransactionProcessedGroupID,
		PartitionWatchInterval: time.Duration(cfg.KafkaTransactionProcessedPartitionWatchInterval) * time.Millisecond,
		Brokers:                globalCFG.KafkaBrokers,
		Topic:                  cfg.KafkaTransactionProcessedTopic,
		MinBytes:               10e2, // 1KB
		MaxBytes:               10e6, // 10MB
		ErrorLogger:            errLogger,
		MaxWait:                time.Duration(cfg.KafkaTransactionProcessedMaxWaitInterval) * time.Millisecond,
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
				cns.lg.ErrorCtx(ctx, "transaction_processed/consumer: fetch message error", zap.Error(err))
			}
		}
	}
}

func (cns *Consumer) processMessage(ctx context.Context) error {
	m, err := cns.reader.FetchMessage(cns.globalCtx)
	if err != nil {
		return fmt.Errorf("transaction_processed/consumer: fetch message error %w", err)
	}

	payload := events.TransactionProcessed{}

	if err := proto.Unmarshal(m.Value, &payload); err != nil {
		return fmt.Errorf("transaction_processed/consumer: unmarshal message error %w", err)
	}

	cns.lg.InfoCtx(ctx, "consumed message", zap.Any("message", &payload))

	processedAt, err := time.Parse(time.RFC3339Nano, payload.ProcessedAt)
	if err != nil {
		return fmt.Errorf("order_created/consumer: unmarshal message error %w", err)
	}

	var operation string
	if payload.Operation == events.TransactionOperations_DEBIT {
		operation = "debit"
	} else {
		operation = "credit"
	}

	if err := cns.events.SaveTransactionCreated(
		ctx,
		&models.TransactionEvent{
			UUID:  payload.EventUuid,
			State: models.OrderEventNewState,
			Name:  repositories.TransactionProcessedEventName,
			Meta: &models.TransactionEventMeta{
				UUID:            payload.Uuid,
				Amount:          payload.Amount,
				AccountID:       payload.AccountId,
				TransactionUUID: payload.Uuid,
				OrderNumber:     payload.OrderNumber,
				Operation:       operation,
				ProcessedAt:     processedAt,
			},
		},
	); err != nil {
		return fmt.Errorf("transaction_processed/consumer: save message error %w", err)
	}

	if err := cns.reader.CommitMessages(ctx, m); err != nil {
		return fmt.Errorf("transaction_processed/consumer: failed to commit messages %w", err)
	}

	return nil
}
