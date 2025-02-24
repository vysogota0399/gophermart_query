package handlers

import (
	"context"
	"time"

	query_orders "github.com/vysogota0399/gophermart_protos/gen/queries/orders"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type OrdersCollectionHandler struct {
	query_orders.UnimplementedQueryOrdersServer

	lg     *logging.ZapLogger
	orders OrdersRepository
}

type OrdersRepository interface {
	SerchByAccoutID(ctx context.Context, accountID int64) ([]*models.Order, error)
}

func NewOrdersCollectionHandler(orders OrdersRepository, lg *logging.ZapLogger) *OrdersCollectionHandler {
	return &OrdersCollectionHandler{orders: orders, lg: lg}
}

func (h *OrdersCollectionHandler) OrdersCollection(ctx context.Context, params *query_orders.QueryOrdersParams) (*query_orders.QureyOrdersResponse, error) {
	orders, err := h.orders.SerchByAccoutID(ctx, params.AccountId)
	if err != nil {
		h.lg.ErrorCtx(ctx, "search orders failed", zap.Error(err), zap.Any("params", params))
		return nil, status.Errorf(codes.Internal, "search orders failed")
	}

	responseOrders := []*query_orders.Order{}

	for _, order := range orders {
		responseOrders = append(responseOrders, &query_orders.Order{
			State:      order.State,
			Number:     order.Number,
			Accrual:    float64(order.Accrual) / 100,
			Uuid:       order.UUID,
			UploadedAt: order.UploadedAt.Format(time.RFC3339Nano),
		})
	}

	return &query_orders.QureyOrdersResponse{
		Orders: responseOrders,
	}, nil
}
