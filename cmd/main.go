package main

import (
	"context"
	"crabspy"
	"crabspy/sql"
	"crabspy/web"
	"fmt"
	"log"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	cfg := crabspy.LoadSettings()

	db, err := sql.NewDatabase(ctx, cfg.DBPath)
	if err != nil {
		return fmt.Errorf("initialize database: %w", err)
	}
	defer db.Close()

	if err := web.RunBlocking(ctx, db); err != nil {
		return fmt.Errorf("run web server: %w", err)
	}
	return nil
}
