package handlers

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"

	query_accounting "github.com/vysogota0399/gophermart_protos/gen/queries/accounting"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	money "google.golang.org/genproto/googleapis/type/money"
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

func (h GetBalanceHandler) GetBalance(ctx context.Context, params *query_accounting.GetBalanceParams) (*query_accounting.GetBalanceResponse, error) {
	balance, err := h.repository.Balance(ctx, params.Account.Id)
	if err != nil {
		h.lg.ErrorCtx(ctx, "calculate balance failed", zap.Error(err))

		return nil, status.Errorf(codes.Internal, "calculate balance failed")
	}

	return &query_accounting.GetBalanceResponse{
		Balance:   &money.Money{Units: balance.Balance, CurrencyCode: "RUB"},
		Withdrawn: &money.Money{Units: balance.Credit, CurrencyCode: "RUB"},
	}, nil
}
