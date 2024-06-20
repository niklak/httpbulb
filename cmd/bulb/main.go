package main

import (
	"context"
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/niklak/httpbulb"
)

const logPrefix string = "BULB SERVER"

type config struct {
	Host         string        `env:"HOST" envDefault:"localhost"`
	Port         int           `env:"PORT" envDefault:"8080"`
	Addr         string        `env:"ADDR,expand" envDefault:"$HOST:${PORT}"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"120s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"120s"`
	CertPath     string        `env:"CERT_PATH"`
	KeyPath      string        `env:"KEY_PATH"`
}

func getTLSConfig(certPath, keyPath string) (tlsConfig *tls.Config, err error) {
	cert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return
	}
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	return
}

func main() {
	cfg := config{}
	opts := env.Options{Prefix: "SERVER_"}
	if err := env.ParseWithOptions(&cfg, opts); err != nil {
		log.Fatalf("[ERROR] %s: %v\n", logPrefix, err)
	}

	var err error
	stop := make(chan os.Signal, 1)

	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	r := httpbulb.NewRouter(middleware.Logger, middleware.Recoverer)

	srv := &http.Server{
		Addr:         cfg.Addr,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		Handler:      r,
	}

	type serverListenFn func() error

	var listenAndServe serverListenFn

	tlsConfig, err := getTLSConfig(cfg.CertPath, cfg.KeyPath)

	if err != nil {
		log.Printf("[WARNING] %s: can't load TLS certificates: %v\n", logPrefix, err)
	}

	if tlsConfig != nil {
		log.Printf("[INFO] %s: TLS Enabled\n", logPrefix)
		srv.TLSConfig = tlsConfig
		listenAndServe = func() error {
			return srv.ListenAndServeTLS("", "")
		}
	} else {
		listenAndServe = srv.ListenAndServe
	}

	go func() {
		log.Printf("[INFO] %s: START SERVING ON %s\n", logPrefix, cfg.Addr)
		if err := listenAndServe(); err != nil {
			log.Printf("[WARNING] %s: %v\n", logPrefix, err)
		}
	}()

	<-stop
	log.Println("[INFO] SERVER: shutting down...")
	if err = srv.Shutdown(context.Background()); err != nil {
		log.Printf("[ERROR] Server: shutdown %v", err)
	}
	log.Println("[INFO] SERVER: gracefully stopped")
}