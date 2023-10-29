package dbhandlers

import (
	"net/http"

	"github.com/eutjeng/go-musthave-metrics-tpl/internal/server/dbstorage"
	"go.uber.org/zap"
)

// PingHandler checks database connection
func PingHandler(sugar *zap.SugaredLogger, dbStorage dbstorage.StorageInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := dbStorage.Ping()
		if err != nil {
			sugar.Errorf("Database ping failed: %s", err)
			http.Error(w, "Internal Server Error", 500)
			return
		}
		sugar.Info("Successfully pinged the database")

		w.WriteHeader(200)
		_, err = w.Write([]byte("OK"))
		if err != nil {
			sugar.Errorf("Failed to write response: %s", err)
			http.Error(w, "Failed to write response", 500)
		}

		sugar.Info("Successfully sent response")
	}
}
