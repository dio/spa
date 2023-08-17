package main

import (
	"context"
	"embed"
	"flag"
	"net/http"
	"time"

	"github.com/dio/spa"
)

var prefix string

//go:embed client/dist
var clientFS embed.FS

func init() {
	flag.StringVar(&prefix, "prefix", "", "Deployment path e.g. ok")
}

func main() {
	flag.Parse()

	// The client/dist is built using PUBLIC_URL=%DEPLOYMENT_PATH% yarn build. See: example/client/vite.config.ts.
	assets, _ := spa.NewAssets(clientFS, "client/dist", spa.WithPrefix("%DEPLOYMENT_PATH%", prefix, spa.NewInMemAfero()))
	srv := newServer(assets)
	srv.Run(context.Background())
}

func newServer(assets *spa.Assets) *server {
	return &server{
		assets: assets,
	}
}

type server struct {
	assets *spa.Assets
}

func (s *server) Run(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.assets.ServeHTTP)

	srv := &http.Server{
		Addr:              "localhost:3000",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errors := make(chan error, 1)

	go func() {
		errors <- srv.ListenAndServe()
	}()

	var err error
	select {
	case err = <-errors:
	case <-ctx.Done():
	}

	if err != nil {
		return err
	}

	// We set 5s before forcefully canceling the server.
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != http.ErrServerClosed {
		return err
	}

	return nil
}
