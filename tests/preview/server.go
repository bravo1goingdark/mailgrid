// preview/server.go
package preview

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// StartServer starts an HTTP server and shuts down on shutdownChan signal.
func StartServer(html string, port int, shutdownChan <-chan struct{}) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if _, err := fmt.Fprint(w, html); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	go func() {
		<-shutdownChan
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}()

	err := srv.ListenAndServe()
	if err != http.ErrServerClosed {
		return err
	}
	return nil
}
