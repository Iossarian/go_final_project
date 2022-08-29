package main

import (
	"abf/internal/bucket"
	"abf/internal/config"
	"abf/internal/list"
	"abf/internal/server"
	"abf/internal/service"
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config", "./configs/config.yml", "Path to configuration file")
}

func main() {
	flag.Parse()

	cf := config.NewConfig(configPath)
	bs := bucket.NewStorage(cf)
	ls := list.NewStorage(cf)
	abfService := service.NewService(bs, ls)
	grpcServer := server.NewServer(abfService, cf)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	go func() {
		if err := grpcServer.Start(ctx); err != nil {
			log.Printf("fail start server: %e", err)
		}
	}()

	<-ctx.Done()
	grpcServer.Stop()
}
