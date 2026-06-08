package main

import (
	"context"
	"database/sql"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	dbmigrations "github.com/reform-barber/api/db"
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

func migrateDatabase(pool *pgxpool.Pool) {
	db := stdlib.OpenDBFromPool(pool)

	err := runMigrations(db)
	exitOnErr(err)

	logger.Info().Msg("successfully migrated database")
}

func runMigrations(db *sql.DB) error {
	goose.SetBaseFS(dbmigrations.FS)
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}
	return goose.Up(db, "migrations")
}
