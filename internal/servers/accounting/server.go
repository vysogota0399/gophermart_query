package accounting

import (
	"context"
	"fmt"
	"net"

	query_accounting "github.com/vysogota0399/gophermart_protos/gen/queries/accounting"
	"github.com/vysogota0399/gophermart_query/internal/config"
	"github.com/vysogota0399/gophermart_query/internal/logging"
	"github.com/vysogota0399/gophermart_query/internal/repositories"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type Server struct {
	handler query_accounting.QueryAccountingServer
	cfg     *config.Config
	srv     *grpc.Server
}

func (s *Server) Start() error {
	lis, err := net.Listen("tcp", s.cfg.AccountingServerHTTPAddress)
	if err != nil {
		return err
	}

	query_accounting.RegisterQueryAccountingServer(s.srv, s.handler)
	go s.srv.Serve(lis)

	return nil
}

func (s *Server) Stop() {
	s.srv.GracefulStop()
}

func NewServer(h query_accounting.QueryAccountingServer, lc fx.Lifecycle, cfg *config.Config, lg *logging.ZapLogger) *Server {
	srv := &Server{cfg: cfg, srv: grpc.NewServer(), handler: h}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				lg.InfoCtx(
					ctx,
					fmt.Sprintf("start processong %s GRPC srequest", repositories.OrderCreatedEventName),
					zap.Any("config", srv.cfg),
				)

				return srv.Start()
			},
			OnStop: func(ctx context.Context) error {
				srv.Stop()
				return nil
			},
		},
	)

	return srv
}
