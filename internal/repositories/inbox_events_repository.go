package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"github.com/vysogota0399/gophermart_query/internal/storage"
)

var OrderCreatedEventName = "order_created"
var OrderUpdatedEventName = "order_updated"
var TransactionProcessedEventName = "transaction_processed"

type InboxEventsRepository struct {
	strg InboxEventsStorage
	lg   *logging.ZapLogger
}

type InboxEventsStorage interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func NewInboxEventsRepository(strg *storage.Storage, lg *logging.ZapLogger) *InboxEventsRepository {
	return &InboxEventsRepository{strg: strg.DB, lg: lg}
}

func (rep *InboxEventsRepository) ReserveOrderCreatedEvent(ctx context.Context) (*models.OrderCreatedEvent, error) {
	tx, err := rep.strg.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("inbox_events_repository: create tx error %w", err)
	}
	defer tx.Rollback(ctx)

	e := &models.OrderCreatedEvent{Name: OrderCreatedEventName}
	row := tx.QueryRow(
		ctx,
		`
			SELECT uuid, message
			FROM inbox_events
			WHERE name = $1 AND state = $2
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`,
		OrderCreatedEventName, models.OrderEventNewState)

	if err := row.Scan(&e.UUID, &e.Meta); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("invox_events_repository scan attributes error %w", err)
	}

	if err := rep.setStateTX(ctx, e.UUID, models.OrderEventProcessingState, tx); err != nil {

		return nil, fmt.Errorf("invox_events_repository set new stateerror %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("invox_events_repository commit tx error %w", err)
	}

	return e, nil
}

func (rep *InboxEventsRepository) ReserveTransactionProcessedEvent(ctx context.Context) (*models.TransactionEvent, error) {
	tx, err := rep.strg.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("inbox_events_repository: create tx error %w", err)
	}
	defer tx.Rollback(ctx)

	e := &models.TransactionEvent{Name: TransactionProcessedEventName}
	row := tx.QueryRow(
		ctx,
		`
			SELECT uuid, message
			FROM inbox_events
			WHERE name = $1 AND state = $2
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			LIMIT 1
		`,
		TransactionProcessedEventName, models.OrderEventNewState)

	if err := row.Scan(&e.UUID, &e.Meta); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("invox_events_repository select models.Event error %w", err)
	}

	if err := rep.setStateTX(ctx, e.UUID, models.OrderEventProcessingState, tx); err != nil {
		return nil, fmt.Errorf("invox_events_repository set new stateerror %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("invox_events_repository commit tx error %w", err)
	}

	return e, nil
}

func (rep *InboxEventsRepository) ReserveOrderUpdatedEvent(ctx context.Context) (*models.OrderUpdatedEvent, error) {
	tx, err := rep.strg.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("inbox_events_repository: create tx error %w", err)
	}
	defer tx.Rollback(ctx)

	e := &models.OrderUpdatedEvent{Name: OrderUpdatedEventName}
	row := tx.QueryRow(
		ctx,
		`
			SELECT uuid, message
			FROM inbox_events
			WHERE name = $1 AND state = $2
			ORDER BY created_at ASC
			FOR UPDATE SKIP LOCKED
			limit 1
		`,
		OrderUpdatedEventName, models.OrderEventNewState)

	if err := row.Scan(&e.UUID, &e.Meta); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}

		return nil, fmt.Errorf("invox_events_repository select models.Event error %w", err)
	}

	if err := rep.setStateTX(ctx, e.UUID, models.OrderEventProcessingState, tx); err != nil {
		return nil, fmt.Errorf("invox_events_repository set new stateerror %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("invox_events_repository commit tx error %w", err)
	}

	return e, nil
}

func (rep *InboxEventsRepository) SetState(ctx context.Context, uuid string, newState string) error {
	tx, err := rep.strg.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("inbox_events_repository: create tx error %w", err)
	}
	defer tx.Rollback(ctx)

	if err := rep.setStateTX(ctx, uuid, newState, tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (rep *InboxEventsRepository) setStateTX(ctx context.Context, uuid string, newState string, tx pgx.Tx) error {
	if _, err := tx.Exec(ctx,
		`
			UPDATE inbox_events
			SET state = $1
			WHERE uuid = $2
		`,
		newState, uuid); err != nil {
		return fmt.Errorf("invox_events_repository: state error %w", err)
	}

	return nil
}

func (rep *InboxEventsRepository) SaveOrderCreated(ctx context.Context, in *models.OrderCreatedEvent) error {
	_, err := rep.strg.Exec(
		ctx,
		`
			INSERT INTO inbox_events(uuid, state, name, message)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`,
		in.UUID, in.State, in.Name, in.Meta,
	)

	if err != nil {
		return fmt.Errorf("invox_events_repository: save event error %w", err)
	}

	return nil
}

func (rep *InboxEventsRepository) SaveOrderUpdated(ctx context.Context, in *models.OrderUpdatedEvent) error {
	_, err := rep.strg.Exec(
		ctx,
		`
			INSERT INTO inbox_events(uuid, state, name, message)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`,
		in.UUID, in.State, in.Name, in.Meta,
	)

	if err != nil {
		return fmt.Errorf("invox_events_repository: save evebt error %w", err)
	}

	return nil
}

func (rep *InboxEventsRepository) SaveTransactionCreated(ctx context.Context, in *models.TransactionEvent) error {
	_, err := rep.strg.Exec(
		ctx,
		`
			INSERT INTO inbox_events(uuid, state, name, message)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT DO NOTHING
		`,
		in.UUID, in.State, in.Name, in.Meta,
	)

	if err != nil {
		return fmt.Errorf("invox_events_repository: save event error %w", err)
	}

	return nil
}