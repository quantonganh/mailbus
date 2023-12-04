package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"

	"github.com/quantonganh/mailbus"
	"github.com/quantonganh/mailbus/bolt"
	"github.com/quantonganh/mailbus/gmail"
	"github.com/quantonganh/mailbus/http"
)

func main() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatal(err)
	}

	viper.SetDefault("http.addr", ":8080")

	var config *mailbus.Config
	if err := viper.Unmarshal(&config); err != nil {
		log.Fatal(err)
	}

	if err := sentry.Init(sentry.ClientOptions{
		Dsn: config.Sentry.DSN,
	}); err != nil {
		log.Fatalf("sentry.Init: %v", err)
	}
	defer sentry.Flush(2 * time.Second)

	a := newApp(config)

	ctx, cancel := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		cancel()
	}()

	if err := a.Run(ctx); err != nil {
		_ = a.Close()
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	<-ctx.Done()

	if err := a.Close(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

type app struct {
	config     *mailbus.Config
	db         *bolt.DB
	httpServer *http.Server
}

func newApp(config *mailbus.Config) *app {
	httpServer, err := http.NewServer()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	return &app{
		config:     config,
		db:         bolt.NewDB(config.DB.Path),
		httpServer: httpServer,
	}
}

func (a *app) Run(ctx context.Context) error {
	if err := a.db.Open(); err != nil {
		return err
	}

	a.httpServer.Addr = a.config.HTTP.Addr

	if err := a.httpServer.Open(); err != nil {
		return err
	}

	a.httpServer.SubscriptionService = bolt.NewSubscriptionService(a.db)
	a.httpServer.NewsletterService = gmail.NewNewsletterService(a.config, a.httpServer.URL(), a.httpServer.SubscriptionService)

	return nil
}

func (a *app) Close() error {
	if a.httpServer != nil {
		if a.httpServer.NewsletterService != nil {
			if err := a.httpServer.NewsletterService.Stop(); err != nil {
				return err
			}
		}

		if err := a.httpServer.Close(); err != nil {
			return err
		}
	}

	if a.db != nil {
		if err := a.db.Close(); err != nil {
			return err
		}
	}

	return nil
}
