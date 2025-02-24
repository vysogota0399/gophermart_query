package storage

import (
	"context"
	"embed"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/vysogota0399/gophermart_query/internal/config"
	"go.uber.org/fx"
)

type Storage struct {
	DB *pgxpool.Pool
}

func NewStorage(lc fx.Lifecycle, cfg *config.Config) (*Storage, error) {
	dbcfg, err := pgxpool.ParseConfig(cfg.DatabaseDSN)
	if err != nil {
		return nil, err
	}

	dbcfg.MaxConns = 10

	dbpool, err := pgxpool.NewWithConfig(context.Background(), dbcfg)
	if err != nil {
		return nil, err
	}

	strg := &Storage{DB: dbpool}

	lc.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				if err := dbpool.Ping(ctx); err != nil {
					return err
				}

				return strg.RunMigration()
			},
			OnStop: func(ctx context.Context) error {
				strg.DB.Close()
				return nil
			},
		},
	)

	return strg, nil
}

//go:embed migrations/*.sql
var embedMigrations embed.FS

func (s *Storage) RunMigration() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect(string(goose.DialectPostgres)); err != nil {
		return err
	}

	return goose.Up(stdlib.OpenDBFromPool(s.DB), "migrations")
}
