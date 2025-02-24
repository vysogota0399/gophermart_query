package main

import (
	main_config "github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"github.com/vysogota0399/gophermart_query/internal/storage"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox/order_updated"
	"go.uber.org/fx"
)

func main() {
	fx.New(CreateApp()).Run()
}

func CreateApp() fx.Option {
	return fx.Options(
		fx.Provide(
			logging.NewZapLogger,
			logging.NewKafkaErrorLogger,
			logging.NewKafkaLogger,
			storage.NewStorage,

			order_updated.NewDaemon,
			order_updated.NewConsumer,
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(order_updated.InboxEventsRepository))),
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(order_updated.ConsumerInboxEventsRepository))),
			fx.Annotate(repositories.NewOrdersRepository, fx.As(new(order_updated.OrdersRepository))),
		),
		fx.Supply(main_config.MustNewConfig(), transaction_inbox.MustNewConfig()),
		fx.Invoke(startDaemon, startConsumer),
	)
}

func startDaemon(*order_updated.Daemon)     {}
func startConsumer(*order_updated.Consumer) {}
