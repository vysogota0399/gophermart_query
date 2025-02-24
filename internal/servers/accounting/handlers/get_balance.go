package handlers

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	query_accounting "github.com/vysogota0399/gophermart_protos/gen/queries/accounting"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
)

type GetBalanceHandler struct {
	lg         *logging.ZapLogger
	repository BalanceRepository
}

type BalanceRepository interface {
	Balance(ctx context.Context, accountID int64) (*repositories.Balance, error)
}

func NewGetBalanceHandler(repository BalanceRepository, lg *logging.ZapLogger) *GetBalanceHandler {
	return &GetBalanceHandler{lg: lg, repository: repository}
}

func (h GetBalanceHandler) GetBalance(ctx context.Context, params *query_accounting.BalanceParams) (*query_accounting.BalanceResponse, error) {
	balance, err := h.repository.Balance(ctx, params.AccountId)
	if err != nil {
		h.lg.ErrorCtx(ctx, "calculate balance failed", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "calculate balance failed")
	}

	return &query_accounting.BalanceResponse{
		Balance:   float64(balance.Balance) / 100,
		Withdrawn: float64(balance.Credit) /100,
	}, nil
}
