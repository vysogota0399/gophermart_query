package main

import (
	main_config "github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"github.com/vysogota0399/gophermart_query/internal/storage"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox"
	"github.com/vysogota0399/gophermart_query/internal/transaction_inbox/transaction_processed"
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

			transaction_processed.NewDaemon,
			transaction_processed.NewConsumer,
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(transaction_processed.InboxEventsRepository))),
			fx.Annotate(repositories.NewTransactionsRepository, fx.As(new(transaction_processed.TransactionsRepository))),
			fx.Annotate(repositories.NewInboxEventsRepository, fx.As(new(transaction_processed.ConsumerInboxEventsRepository))),
			fx.Annotate(repositories.NewOrdersRepository, fx.As(new(transaction_processed.OrdersRepository))),
		),
		fx.Supply(main_config.MustNewConfig(), transaction_inbox.MustNewConfig()),
		fx.Invoke(startDaemon, startConsumer),
	)
}

func startDaemon(*transaction_processed.Daemon)     {}
func startConsumer(*transaction_processed.Consumer) {}
