package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"go.uber.org/zap"

	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"github.com/vysogota0399/gophermart_query/internal/storage"
)

type OrdersRepository struct {
	strg Storage
	lg   *logging.ZapLogger
}

type Storage interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	Query(ctx context.Context, query string, args ...any) (pgx.Rows, error)
	Exec(ctx context.Context, query string, args ...any) (pgconn.CommandTag, error)
}

func NewOrdersRepository(strg *storage.Storage, lg *logging.ZapLogger) *OrdersRepository {
	return &OrdersRepository{strg: strg.DB, lg: lg}
}

func (rep *OrdersRepository) Create(ctx context.Context, order *models.Order) error {
	_, err := rep.strg.Exec(
		ctx,
		`
		  INSERT INTO ORDERS(uuid, state, number, account_id, uploaded_at)
			VALUES ($1, $2, $3, $4, $5)
		`,
		order.UUID, order.State, order.Number, order.AccountID, order.UploadedAt,
	)

	if err != nil {
		return fmt.Errorf("orders_repository: save order error %w", err)
	}

	return nil
}

func (rep *OrdersRepository) UpdateState(ctx context.Context, in *models.Order) error {
	_, err := rep.strg.Exec(
		ctx,
		`
			UPDATE orders
			SET state = $1
			WHERE uuid = $2
		`,
		in.State, in.UUID,
	)

	if err != nil {
		return fmt.Errorf("orders_repository: update order error %w", err)
	}

	return nil
}

func (rep *OrdersRepository) UpdateAccrualTX(ctx context.Context, in *models.Order, tx pgx.Tx) error {
	rep.lg.DebugCtx(
		ctx,
		"update order accrual query",
		zap.String("order_number", in.Number),
		zap.Int64("accrual", in.Accrual),
	)
	_, err := tx.Exec(
		ctx,
		`
			UPDATE orders
			SET accrual = $1
			WHERE number = $2
		`,
		in.Accrual, in.Number,
	)

	if err != nil {
		return fmt.Errorf("orders_repository: update order error %w", err)
	}

	return nil
}

func (rep *OrdersRepository) SerchByAccoutID(ctx context.Context, accountID int64) ([]*models.Order, error) {
	rows, err := rep.strg.Query(
		ctx,
		`
			SELECT uuid, number, state, accrual, account_id, uploaded_at
			FROM orders
			WHERE account_id = $1
			ORDER BY uploaded_at DESC
		`,
		accountID,
	)
	if err != nil {
		return nil, fmt.Errorf("orders_repository: query orders error %w", err)
	}

	orders := []*models.Order{}
	for rows.Next() {
		order := &models.Order{}
		if err := rows.Scan(
			&order.UUID,
			&order.Number,
			&order.State,
			&order.Accrual,
			&order.AccountID,
			&order.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("orders_repository: scan query orders error %w", err)
		}

		orders = append(orders, order)
	}

	return orders, nil
}
