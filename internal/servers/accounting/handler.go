package accounting

import (
	"context"

	query_accounting "github.com/vysogota0399/gophermart_protos/gen/queries/accounting"
	"github.com/vysogota0399/gophermart_query/internal/servers/accounting/handlers"
)

type Handler struct {
	query_accounting.UnimplementedQueryAccountingServer
	handlers.GetBalanceHandler
	handlers.GetWithdrawalsHandler
}

func NewHandler(
	getBalance *handlers.GetBalanceHandler,
	getWithdrawals *handlers.GetWithdrawalsHandler,
) *Handler {
	return &Handler{GetBalanceHandler: *getBalance, GetWithdrawalsHandler: *getWithdrawals}
}

func (h *Handler) GetWithdrawals(ctx context.Context, params *query_accounting.GetWithdrawalsParams) (*query_accounting.GetWithdrawalsResponse, error) {
	return h.GetWithdrawalsHandler.GetWithdrawals(ctx, params)
}

func (h *Handler) GetBalance(ctx context.Context, params *query_accounting.BalanceParams) (*query_accounting.BalanceResponse, error) {
	return h.GetBalanceHandler.GetBalance(ctx, params)
}
