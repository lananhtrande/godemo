package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	addr := env("ADDR", ":8080")
	allowedOrigin := env("ALLOWED_ORIGIN", "http://localhost:3000")

	store := NewStore([]Site{
		{ID: "wue-nord", Name: "Würzburg Nord", CapacityMWh: 30, MaxPowerMW: 15, SoC: 62, Status: "discharging"},
		{ID: "schweinfurt", Name: "Schweinfurt", CapacityMWh: 17.5, MaxPowerMW: 8, SoC: 35, Status: "charging"},
		{ID: "bamberg", Name: "Bamberg", CapacityMWh: 10, MaxPowerMW: 5, SoC: 80, Status: "idle"},
		{ID: "kitzingen", Name: "Kitzingen", CapacityMWh: 4, MaxPowerMW: 2, SoC: 15, Status: "charging"},
	})

	// Stop cleanly on Ctrl+C or when Kubernetes/Docker sends SIGTERM.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go runSimulation(ctx, store)

	srv := &http.Server{
		Addr:              addr,
		Handler:           (&API{store: store, allowedOrigin: allowedOrigin}).routes(),
		ReadHeaderTimeout: 5 * time.Second,
		// No WriteTimeout: SSE responses stay open for a long time.
	}

	go func() {
		log.Printf("axos-demo backend listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

// runSimulation updates all sites once per second.
// One real second = one minute on site, so the SoC visibly changes.
func runSimulation(ctx context.Context, store *Store) {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			store.Step(time.Minute, rng)
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
