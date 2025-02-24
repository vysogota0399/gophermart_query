package handlers

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	query_accounting "github.com/vysogota0399/gophermart_protos/gen/queries/accounting"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
)

type GetWithdrawalsHandler struct {
	lg         *logging.ZapLogger
	repository WithdrawalsRepository
}

type WithdrawalsRepository interface {
	Debits(ctx context.Context, accountID int64) ([]*models.Transaction, error)
}

func NewGetWithdrawalsHandler(repository WithdrawalsRepository, lg *logging.ZapLogger) *GetWithdrawalsHandler {
	return &GetWithdrawalsHandler{repository: repository, lg: lg}
}

func (h GetWithdrawalsHandler) GetWithdrawals(
	ctx context.Context,
	params *query_accounting.GetWithdrawalsParams,
) (*query_accounting.GetWithdrawalsResponse, error) {
	debits, err := h.repository.Debits(ctx, params.AccountId)
	if err != nil {
		h.lg.ErrorCtx(ctx, "get withdrawals failed", zap.Error(err))
		return nil, status.Errorf(codes.Internal, "get withdrawals failed")
	}

	withdrawals := []*query_accounting.Withdrawal{}
	for _, d := range debits {
		withdrawals = append(
			withdrawals,
			&query_accounting.Withdrawal{
				OrderNumber: d.OrderNumber,
				Sum:         float64(d.Amount)/100,
				ProcessedAt: d.ProcessedAt.Format(time.RFC3339Nano),
			},
		)
	}

	return &query_accounting.GetWithdrawalsResponse{
		Withdrawals: withdrawals,
	}, nil
}
