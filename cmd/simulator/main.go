// Command simulator is a standalone entry point for the ISPVisualMonitor telemetry
// simulator. It connects to the application database and generates realistic ISP
// metrics, allowing dashboards, alerting, and incident workflows to be exercised
// without a real agent or MikroTik devices.
//
// Usage:
//
//	go run ./cmd/simulator [flags]
//
// Environment variables:
//
//	DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME  — database connection
//	SIM_MODE          — deterministic | scenario | random (default: scenario)
//	SIM_SEED          — RNG seed for deterministic mode (default: 42)
//	SIM_INTERVAL      — generation interval, e.g. "30s" (default: 30s)
//	SIM_SCENARIO      — initial scenario name (default: healthy)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/database"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/internal/simulator"
	"github.com/MohamadKhaledAbbas/ISPVisualMonitor/pkg/config"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("[sim] ")

	// --- database connection (reuse app config) ---
	dbCfg := config.DatabaseConfig{
		Host:     envStr("DB_HOST", "localhost"),
		Port:     envInt("DB_PORT", 5432),
		User:     envStr("DB_USER", "ispmonitor"),
		Password: envStr("DB_PASSWORD", "ispmonitor"),
		DBName:   envStr("DB_NAME", "ispmonitor"),
		SSLMode:  envStr("DB_SSLMODE", "disable"),
		MaxConns: 5,
		MinConns: 1,
	}

	db, err := database.NewConnection(dbCfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// --- simulator config ---
	simCfg := simulator.DefaultConfig()
	simCfg.Mode = parseMode(envStr("SIM_MODE", "scenario"))
	simCfg.Seed = int64(envInt("SIM_SEED", 42))
	simCfg.Interval = envDuration("SIM_INTERVAL", 30*time.Second)
	simCfg.Scenario = simulator.ScenarioName(envStr("SIM_SCENARIO", "healthy"))

	// If a scenario was provided as a CLI arg, use it (e.g. go run ./cmd/simulator router-down)
	if len(os.Args) > 1 {
		arg := strings.TrimSpace(os.Args[1])
		if arg != "" && arg[0] != '-' {
			simCfg.Scenario = simulator.ScenarioName(arg)
		}
	}

	svc := simulator.NewService(db.DB, simCfg)
	log.Printf("Simulator configured: %s", svc)

	fmt.Println()
	fmt.Println("Available scenarios:")
	for _, s := range simulator.AllScenarios() {
		marker := "  "
		if s == simCfg.Scenario {
			marker = "▸ "
		}
		fmt.Printf("  %s%s\n", marker, s)
	}
	fmt.Println()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Received shutdown signal")
		cancel()
	}()

	if err := svc.Start(ctx); err != nil {
		log.Fatalf("Simulator error: %v", err)
	}

	log.Println("Simulator stopped")
}

// --- helpers ---

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func envDuration(key string, def time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}

func parseMode(s string) simulator.SimMode {
	switch strings.ToLower(s) {
	case "deterministic":
		return simulator.ModeDeterministic
	case "random":
		return simulator.ModeRandom
	default:
		return simulator.ModeScenario
	}
}
