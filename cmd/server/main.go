package main

import (
	"context"
	"log"
	"sync"

	_ "github.com/lib/pq"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/appinit"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/signalhandlers"
)

func main() {
	var wg sync.WaitGroup

	cfg, sugar, syncFunc, err := appinit.InitServerApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %s", err)
	}
	defer syncFunc()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	store, errInit := appinit.InitStore(ctx, cfg, sugar, &wg)
	if errInit != nil {
		sugar.Fatalf("Failed to initialize storage: %v", errInit)
	}

	srv := appinit.InitServer(cfg, router.SetupRouter(ctx, cfg, sugar, store, cfg.StoreInterval == 0))

	quitChan, signalChan := appinit.InitSignalHandling()

	go signalhandlers.HandleSignals(signalChan, quitChan, store, cfg, sugar)

	errChan := make(chan error)
	appinit.StartServer(sugar, srv, errChan)

	go func() {
		signalhandlers.HandleServerErrors(errChan, sugar, cfg)
	}()

	wg.Add(1)
	go func() {
		signalhandlers.HandleShutdownServer(ctx, quitChan, srv, sugar, &wg, cancel)

	}()
	wg.Wait()
}
