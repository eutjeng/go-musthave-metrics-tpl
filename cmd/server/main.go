package main

import (
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/appinit"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/filestorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/signalhandlers"
)

func main() {
	// initialization: config, logger, storage, and server
	cfg, sugar, syncFunc := appinit.InitApp()
	defer syncFunc()
	storage := storage.NewInMemoryStorage()
	srv := appinit.InitServer(cfg, router.SetupRouter(sugar, storage, cfg.StoreInterval == 0))

	// restore data from file if necessary
	filestorage.RestoreData(sugar, storage, cfg)

	// initialize data save mechanisms
	appinit.InitDataSave(storage, cfg, sugar)

	// initialize signal handling
	quitChan, signalChan := appinit.InitSignalHandling()
	go signalhandlers.HandleSignals(signalChan, quitChan, storage, cfg, sugar)

	// start server
	errChan := make(chan error)
	appinit.StartServer(srv, errChan)

	// prepare for shutdown or errors
	doneChan := make(chan struct{})

	go func() {
		signalhandlers.HandleServerErrors(errChan, sugar, cfg)
		doneChan <- struct{}{}
	}()

	go func() {
		signalhandlers.HandleShutdownServer(quitChan, srv, sugar)
		doneChan <- struct{}{}
	}()

	// wait for shutdown or error signal
	<-doneChan
}
