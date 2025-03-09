package handlers

import (
	"context"

	"github.com/vysogota0399/gophermart_protos/gen/common"
	"github.com/vysogota0399/gophermart_protos/gen/entities"
	query_orders "github.com/vysogota0399/gophermart_protos/gen/queries/orders"
	"github.com/vysogota0399/gophermart_protos/utils/amount"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/models"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
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

func (h *OrdersCollectionHandler) OrdersCollection(ctx context.Context, params *query_orders.QueryOrdersRequest) (*query_orders.QureyOrdersResponse, error) {
	orders, err := h.orders.SerchByAccoutID(ctx, params.Account.Id)
	if err != nil {
		h.lg.ErrorCtx(ctx, "search orders failed", zap.Error(err), zap.Any("params", params))
		return nil, status.Errorf(codes.Internal, "search orders failed")
	}

	responseOrders := []*entities.Order{}

	for _, order := range orders {
		responseOrders = append(responseOrders, &entities.Order{
			State:      entities.OrderStates(order.State),
			Number:     order.Number,
			Accrual:    amount.FromInt64(order.Accrual).Money,
			Uuid:       &common.Uuid{Value: order.UUID},
			UploadedAt: timestamppb.New(order.UploadedAt),
		})
	}

	return &query_orders.QureyOrdersResponse{
		Orders: responseOrders,
	}, nil
}
