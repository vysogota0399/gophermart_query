package repositories

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"github.com/vysogota0399/gophermart_query/internal/storage"
)

type TransactionsRepository struct {
	strg TransactionsStorage
	lg   *logging.ZapLogger
}

var TransactionDebit = "debit"
var TransactionCredit = "credit"

type TransactionsStorage interface {
	BeginTx(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error)
	QueryRow(ctx context.Context, query string, args ...any) pgx.Row
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func NewTransactionsRepository(strg *storage.Storage, lg *logging.ZapLogger) *TransactionsRepository {
	return &TransactionsRepository{strg: strg.DB, lg: lg}
}

func (rep *TransactionsRepository) Create(ctx context.Context, in *models.Transaction, tx pgx.Tx) error {
	_, err := tx.Exec(
		ctx,
		`
		  INSERT INTO transactions(uuid, order_number, account_id, amount, operation, processed_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`,
		in.UUID, in.OrderNumber, in.AccountID, in.Amount, in.Operation, in.ProcessedAt,
	)
	if err != nil {
		return fmt.Errorf("transactions_repository: create transactions record error %w", err)
	}

	return nil
}

type Balance struct {
	Credit  int64
	Debit   int64
	Balance int64
}

func (rep *TransactionsRepository) Debits(ctx context.Context, accountID int64) ([]*models.Transaction, error) {
	tx, err := rep.BeginTX(ctx, pgx.TxOptions{})
	if err != nil {
		return nil, fmt.Errorf("transactions_repository: create tx error %w", err)
	}
	defer tx.Rollback(ctx)

	debits := []*models.Transaction{}
	rows, err := tx.Query(
		ctx,
		`
			SELECT uuid, order_number, account_id, amount, processed_at
			FROM transactions
			WHERE account_id = $1 AND operation = $2
		`,
		accountID, TransactionCredit,
	)
	if err != nil {
		return nil, fmt.Errorf("transactions_repository: fetch transactions error %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		debit := models.Transaction{Operation: TransactionDebit}
		if err := rows.Scan(
			&debit.UUID,
			&debit.OrderNumber,
			&debit.AccountID,
			&debit.Amount,
			&debit.ProcessedAt,
		); err != nil {
			return nil, fmt.Errorf("transactions_repository: scan transactions error %w", err)
		}

		debits = append(debits, &debit)
	}

	tx.Commit(ctx)
	return debits, nil
}

func (rep *TransactionsRepository) Balance(ctx context.Context, accountID int64) (*Balance, error) {
	tx, err := rep.BeginTX(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return nil, fmt.Errorf("transactions_repository: create tx error %w", err)
	}
	defer tx.Rollback(ctx)

	debit, err := rep.totalAmmountForAccount(ctx, tx, accountID, TransactionDebit)
	if err != nil {
		return nil, fmt.Errorf("transactions_repository: calc debit error %w", err)
	}

	credit, err := rep.totalAmmountForAccount(ctx, tx, accountID, TransactionCredit)
	if err != nil {
		return nil, fmt.Errorf("transactions_repository: calc debit error %w", err)
	}

	tx.Commit(ctx)
	return &Balance{
		Credit:  credit,
		Debit:   debit,
		Balance: debit - credit,
	}, nil
}

func (rep *TransactionsRepository) totalAmmountForAccount(
	ctx context.Context,
	tx pgx.Tx,
	accountID int64,
	operation string,
) (int64, error) {

	row := tx.QueryRow(
		ctx,
		`
			WITH rows AS ( 
				SELECT amount
				FROM transactions
				WHERE account_id = $1 AND operation = $2
		  )

			SELECT COALESCE(sum(rows.amount), 0) FROM rows;
			`,
		accountID, operation,
	)

	var result int64
	if err := row.Scan(&result); err != nil {
		return 0, fmt.Errorf("transactions_repository: calc sum error %w", err)
	}

	return result, nil
}

func (rep *TransactionsRepository) BeginTX(ctx context.Context, opts pgx.TxOptions) (pgx.Tx, error) {
	return rep.strg.BeginTx(ctx, opts)
}

func (rep *TransactionsRepository) CommitTX(ctx context.Context, tx pgx.Tx) error {
	return tx.Commit(ctx)
}

func (rep *TransactionsRepository) RollbackTX(ctx context.Context, tx pgx.Tx) error {
	return tx.Rollback(ctx)
}
