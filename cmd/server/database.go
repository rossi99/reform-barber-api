package main

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func initDatabaseConn(ctx context.Context) *pgxpool.Pool {
	dsn := mustEnv("DATABASE_URL")

	pool, err := pgxpool.New(ctx, dsn)
	exitOnErr(err)

	err = pool.Ping(context.Background())
	exitOnErr(err)

	logger.Info().Msg("successfully connected to database")
	return pool
}
