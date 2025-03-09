package main

import (
	accounting_query "github.com/vysogota0399/gophermart_protos/gen/queries/accounting"
	orders_query "github.com/vysogota0399/gophermart_protos/gen/queries/orders"
	main_config "github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"github.com/vysogota0399/gophermart_query/internal/servers/accounting"
	accounting_service_handler "github.com/vysogota0399/gophermart_query/internal/servers/accounting/handlers"
	"github.com/vysogota0399/gophermart_query/internal/servers/orders"
	orders_service_handlers "github.com/vysogota0399/gophermart_query/internal/servers/orders/handlers"
	"github.com/vysogota0399/gophermart_query/internal/storage"
	"go.uber.org/fx"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			logging.NewZapLogger,
			storage.NewStorage,

			// GRPS servers
			orders.NewServer,
			fx.Annotate(orders_service_handlers.NewOrdersCollectionHandler, fx.As(new(orders_query.QueryOrdersServer))),
			fx.Annotate(repositories.NewOrdersRepository, fx.As(new(orders_service_handlers.OrdersRepository))),

			accounting.NewServer,
			fx.Annotate(accounting.NewHandler, fx.As(new(accounting_query.QueryAccountingServer))),
			accounting_service_handler.NewGetBalanceHandler,
			fx.Annotate(repositories.NewTransactionsRepository, fx.As(new(accounting_service_handler.BalanceRepository))),
			accounting_service_handler.NewGetWithdrawalsHandler,
			fx.Annotate(repositories.NewTransactionsRepository, fx.As(new(accounting_service_handler.WithdrawalsRepository))),
		),
		fx.Supply(
			main_config.MustNewConfig(),
		),
		fx.Invoke(
			startOrdersServer,
			startAccountingServer,
		),
	)
}

func startOrdersServer(*orders.Server)         {}
func startAccountingServer(*accounting.Server) {}
