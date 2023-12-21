package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/spf13/viper"

	"github.com/quantonganh/mailbus"
	"github.com/quantonganh/mailbus/bolt"
	"github.com/quantonganh/mailbus/gmail"
	"github.com/quantonganh/mailbus/http"
	"github.com/quantonganh/mailbus/sqlite"
)

type DatabaseType string

const (
	BoltDB   DatabaseType = "bolt"
	SQLiteDB DatabaseType = "sqlite"
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
	db         mailbus.Database
	httpServer *http.Server
}

func newApp(config *mailbus.Config) *app {
	db, subscriptionSvc, err := newDatabaseService(DatabaseType(config.DB.Type), config.DB.Path)
	if err != nil {
		log.Fatal(err)
	}

	httpServer, err := http.NewServer()
	if err != nil {
		log.Fatalf("%+v\n", err)
	}
	httpServer.SubscriptionService = subscriptionSvc

	return &app{
		config:     config,
		db:         db,
		httpServer: httpServer,
	}
}

func newDatabaseService(dbType DatabaseType, path string) (mailbus.Database, mailbus.SubscriptionService, error) {
	var (
		db              mailbus.Database
		subscriptionSvc mailbus.SubscriptionService
		err             error
	)

	if dbType == "" {
		dbType = SQLiteDB
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, nil, err
	}

	switch dbType {
	case BoltDB:
		db = bolt.NewDB(path)
		boltDB, ok := db.(*bolt.DB)
		if ok {
			subscriptionSvc = bolt.NewSubscriptionService(boltDB)
		} else {
			err = fmt.Errorf("failed to create BoltDB")
		}
	case SQLiteDB:
		db = sqlite.NewDB(path)
		sqliteDB, ok := db.(*sqlite.DB)
		if ok {
			subscriptionSvc = sqlite.NewSubscriptionService(sqliteDB)
		} else {
			err = fmt.Errorf("failed to create SQLiteDB")
		}
	default:
		err = fmt.Errorf("unsupported database type: %s", dbType)

	}

	return db, subscriptionSvc, err
}

func (a *app) Run(ctx context.Context) error {
	if err := a.db.Open(); err != nil {
		return err
	}

	a.httpServer.Addr = a.config.HTTP.Addr

	if err := a.httpServer.Open(); err != nil {
		return err
	}

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
