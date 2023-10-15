package main

import (
	"log"

	_ "github.com/lib/pq"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/appinit"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/filestorage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/router"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/storage"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/signalhandlers"
)

func main() {
	cfg, sugar, syncFunc, err := appinit.InitApp()
	if err != nil {
		log.Fatalf("Failed to initialize app: %s", err)
	}
	defer syncFunc()

	storage := storage.NewInMemoryStorage()
	dbStorage, err := dbstorage.NewDBStorage(cfg.DbDSN)
	if err != nil {
		sugar.Fatalf("Failed to connect to database: %v", err)
	}

	srv := appinit.InitServer(cfg, router.SetupRouter(sugar, storage, dbStorage, cfg.StoreInterval == 0))

	filestorage.RestoreData(sugar, storage, cfg)
	appinit.InitDataSave(sugar, storage, cfg)

	quitChan, signalChan := appinit.InitSignalHandling()
	go signalhandlers.HandleSignals(signalChan, quitChan, storage, cfg, sugar)
	errChan := make(chan error)

	appinit.StartServer(srv, errChan)

	doneChan := make(chan struct{})
	go func() {
		signalhandlers.HandleServerErrors(errChan, sugar, cfg)
		doneChan <- struct{}{}
	}()
	go func() {
		signalhandlers.HandleShutdownServer(quitChan, srv, sugar)
		doneChan <- struct{}{}
	}()
	<-doneChan
}
