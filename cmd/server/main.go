package main

import (
	"context"
	"log"
	"sync"

	_ "github.com/lib/pq"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/appinit"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/models"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/signalhandlers"
)

func main() {
	var wg sync.WaitGroup

	cfg, sugar, syncFunc, err := appinit.InitApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %s", err)
	}
	defer syncFunc()

	var store models.GeneralStorageInterface
	var errInit error

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if cfg.DBDSN != "" {
		wg.Add(1)
		store, errInit = appinit.InitDBStorage(ctx, cfg, sugar, &wg)
	} else {
		store, errInit = appinit.InitInMemoryStorage(cfg, sugar)
	}

	if errInit != nil {
		sugar.Fatalf("Failed to initialize storage: %v", errInit)
	}

	srv := appinit.InitServer(cfg, router.SetupRouter(sugar, store, cfg.StoreInterval == 0))

	quitChan, signalChan := appinit.InitSignalHandling()

	go signalhandlers.HandleSignals(signalChan, quitChan, store, cfg, sugar)

	errChan := make(chan error)
	appinit.StartServer(srv, errChan)

	go func() {
		signalhandlers.HandleServerErrors(errChan, sugar, cfg)
	}()

	wg.Add(1)
	go func() {
		signalhandlers.HandleShutdownServer(ctx, quitChan, srv, sugar, &wg, cancel)

	}()
	wg.Wait()
}
