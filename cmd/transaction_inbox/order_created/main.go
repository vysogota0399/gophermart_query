package main

import (
	main_config "github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"github.com/vysogota0399/gophermart_query/internal/storage"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox/order_created"
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

			order_created.NewDaemon,
			order_created.NewConsumer,
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(order_created.InboxEventsRepository))),
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(order_created.ConsumerInboxEventsRepository))),
			fx.Annotate(repositories.NewOrdersRepository, fx.As(new(order_created.OrdersRepository))),
		),
		fx.Supply(main_config.MustNewConfig(), transaction_inbox.MustNewConfig()),
		fx.Invoke(startDaemon, startConsumer),
	)
}

func startDaemon(*order_created.Daemon)     {}
func startConsumer(*order_created.Consumer) {}
