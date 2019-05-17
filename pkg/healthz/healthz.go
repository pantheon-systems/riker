package healthz

// Package that implements a healthz healthcheck server.
//
// Adapted from @kelseyhightower:
// - https://github.com/kelseyhightower/app-healthz
// - https://vimeo.com/173610242
//
// Further adapted from
// https://github.com/pantheon-systems/policy-docs/blob/master/pkg/healthz/healthz.go
// to suit our needs
//
// Add a `Healthz()` function to your application components,
// and then register them with this package along with a type
// and description by adding them to `Providers`. This package
// creates a HTTP server that runs all the registered handlers and
// returns any errors.
//
// This server does not use TLS, as most applications already
// run their own TLS servers. One approach is to not expose this
// server's port directly (except for Kube's liveness), and rather
// call the healthz handler internally from a TLS handler that you
// add to your main TLS server.

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
)

type HealthCheckable interface {
	HealthZ() error
}

type ProviderInfo struct {
	Check       HealthCheckable
	Description string
	Type        string
}

type Error struct {
	Type        string
	ErrMsg      string
	Description string
}

type HTTPResponse struct {
	Errors   []Error
	Hostname string
}

type Config struct {
	BindPort  int
	BindAddr  string
	Hostname  string
	Providers []ProviderInfo
	Logger    *logrus.Logger
}

type HealthChecker struct {
	Providers []ProviderInfo
	Server    *http.Server
	Hostname  string
	Logger    *logrus.Logger
}

func New(config Config) (*HealthChecker, error) {
	// Hostname is sent in check results, so that we can tell which pod the health check is failing on.
	log := config.Logger
	if config.Hostname == "" {
		var err error
		config.Hostname, err = os.Hostname()
		if err != nil {
			return nil, fmt.Errorf("could not detect hostname: %s", err.Error())
		}
		log.Info("autodetected hostname as: ", config.Hostname)
	}

	h := &HealthChecker{
		Providers: config.Providers,
		Hostname:  config.Hostname,
	}
	mux := http.NewServeMux()
	mux.Handle("/healthz", http.HandlerFunc(h.HandleHealthz))
	mux.Handle("/liveness", http.HandlerFunc(h.HandleLiveness))
	h.Server = &http.Server{
		Addr:           fmt.Sprintf("%s:%d", config.BindAddr, config.BindPort),
		ReadTimeout:    time.Second * 45,
		WriteTimeout:   time.Second * 45,
		MaxHeaderBytes: 1 << 20,
		Handler:        mux,
	}
	h.Logger = log
	return h, nil
}

// HandleHealthz is the http handler for `/healthz`
func (h *HealthChecker) HandleHealthz(w http.ResponseWriter, r *http.Request) {
	resp := &HTTPResponse{
		Hostname: h.Hostname,
	}

	// Check all our health providers
	for _, provider := range h.Providers {
		err := provider.Check.HealthZ()
		if err != nil {
			resp.Errors = append(resp.Errors, Error{
				Type:        provider.Type,
				ErrMsg:      err.Error(),
				Description: provider.Description,
			})
		}
	}

	if len(resp.Errors) > 0 {
		for _, e := range resp.Errors {
			h.Logger.WithFields(logrus.Fields{
				"error":       e.ErrMsg,
				"healthzDesc": e.Description,
				"healthzType": e.Type,
			}).Error("Check failed")
		}
		w.WriteHeader(http.StatusServiceUnavailable)
	} else {
		h.Logger.Debug("All checks passed")
		w.WriteHeader(http.StatusOK)
	}

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")

	err := enc.Encode(resp)
	if err != nil {
		h.Logger.Error(err)
	}
}

// HandleLiveness is the http handler for `/liveness`
func (h *HealthChecker) HandleLiveness(w http.ResponseWriter, r *http.Request) {
	// log.Debug("Liveness check: OK")
	_, err := w.Write([]byte("OK"))
	if err != nil {
		h.Logger.Errorln(err)
	}
}

// StartHealthz should be run in a new goroutine.
func (h *HealthChecker) StartHealthz() {
	h.Logger.Debug("Starting healthz server")
	err := h.Server.ListenAndServe()
	if err != nil {
		h.Logger.Error(err)
	}
}
