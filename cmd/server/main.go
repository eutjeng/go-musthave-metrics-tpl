package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/handlers"
	"github.com/eutjeng/go-musthave-metrics-tpl/internal/storage"
)

func main() {

	var storage storage.MetricStorage = storage.NewInMemoryStorage()
	mux := http.NewServeMux()

	mux.HandleFunc("/update/", handlers.HandleUpdateMetric(storage))

	go func() {
		for {
			fmt.Print("\033[H\033[2J")
			fmt.Println(storage)
			time.Sleep(1 * time.Second)
		}
	}()

	http.ListenAndServe(":8080", mux)

}
