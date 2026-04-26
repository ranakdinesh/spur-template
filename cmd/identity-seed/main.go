package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spurbase/spur/internal/app"
	"github.com/spurbase/spur/internal/modules/identity/seed"
	"github.com/spurbase/spur/internal/platform/config"
	"github.com/spurbase/spur/internal/platform/db"
)

func main() {
	// 1. Parse Command Line Flags
	migrate := flag.Bool("migrate", false, "Run database migrations")
	seedDB := flag.Bool("seed", false, "Seed the database with initial data")
	migrationPath := flag.String("migration-path", "internal/modules/identity/sql/migrations", "Path to migration files")
	flag.Parse()

	// 2. Load Config (Need DatabaseURL)
	var cfg config.Config
	if err := config.Load(&cfg); err != nil {
		fmt.Printf("Fatal: failed to load config: %v\n", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 3. Execute Migrations (if requested)
	if *migrate {
		fmt.Println("🚀 Running database migrations...")
		if err := db.RunMigrations(cfg.DatabaseURL, *migrationPath); err != nil {
			fmt.Printf("❌ Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ Migrations completed successfully.")
	}

	// 4. Execute Seeding (if requested)
	if *seedDB {
		fmt.Println("🌱 Seeding database...")
		// We open a temporary connection pool just for seeding
		pool := db.NewPool(ctx, cfg.DatabaseURL)

		if err := seed.Run(ctx, pool); err != nil {
			pool.Close()
			fmt.Printf("❌ Seeding failed: %v\n", err)
			os.Exit(1)
		}

		pool.Close()
		fmt.Println("✅ Database seeded successfully.")
		fmt.Println("   - Super Admin: admin@system.local / ChangeMe_123!")
	}

	// If we ran admin tasks, exit here.
	if *migrate || *seedDB {
		return
	}

	// 5. Normal Application Start
	fmt.Println("🔥 Starting Citual Server...")
	application, err := app.New(ctx)
	if err != nil {
		fmt.Printf("Fatal: %v\n", err)
		os.Exit(1)
	}

	// Setup Graceful Shutdown
	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		<-c
		fmt.Println("\nShutting down...")
		cancel()
	}()

	if err := application.Start(ctx); err != nil {
		fmt.Printf("Runtime Error: %v\n", err)
		os.Exit(1)
	}
}
