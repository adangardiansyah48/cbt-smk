package config

import (
	"context"
	"crypto/tls"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ConnectDB(url string) *pgxpool.Pool {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		log.Fatalf("parse DATABASE_URL: %v", err)
	}
	if cfg.ConnConfig.TLSConfig == nil {
		cfg.ConnConfig.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	} else {
		cfg.ConnConfig.TLSConfig.InsecureSkipVerify = true
	}
	cfg.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeSimpleProtocol
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("ping db: %v", err)
	}
	log.Println("database connected")
	return pool
}
