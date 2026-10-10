package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"hexis/internal/config"
	"hexis/internal/router"
)

func main() {
	logger := log.New(os.Stdout, "hexis ", log.LstdFlags|log.LUTC)
	cfg := config.Load()

	gin.SetMode(cfg.Server.GinMode)

	var db *gorm.DB
	if cfg.Database.Enabled {
		var err error
		db, err = gorm.Open(mysql.Open(cfg.Database.DSN()), &gorm.Config{})
		if err != nil {
			logger.Fatal(err)
		}
		sqlDB, err := db.DB()
		if err != nil {
			logger.Fatal(err)
		}
		sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
		sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		err = sqlDB.PingContext(ctx)
		cancel()
		if err != nil {
			logger.Fatal(err)
		}
	}

	server := &http.Server{
		Addr:              cfg.Server.Addr,
		Handler:           router.New(db),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		logger.Printf("listening on %s", server.Addr)
		if err := server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Fatal(err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
	if db != nil {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	}
}
