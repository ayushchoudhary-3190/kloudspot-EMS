package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"EMS/internal/config"
	"EMS/internal/database"
	"EMS/internal/handlers"
	"EMS/internal/repository"
	"EMS/internal/service"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := database.SeedAdmin(db, cfg); err != nil {
		log.Fatalf("failed to seed admin user: %v", err)
	}

	empRepo := repository.NewEmployeeRepository(db)
	adminRepo := repository.NewAdminRepository(db)

	empSvc := service.NewEmployeeService(empRepo)
	authSvc := service.NewAuthService(adminRepo, cfg.JWTSecret, cfg.JWTExpiryHrs)

	router := handlers.NewRouter(cfg, empSvc, authSvc)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("EMS API listening on :%s (docs: http://localhost:%s/swagger)", cfg.Port, cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}
