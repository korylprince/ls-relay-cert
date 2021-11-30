package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
)

func RunServer() error {
	config := new(Config)
	err := envconfig.Process("", config)
	if err != nil {
		return fmt.Errorf("could not process configuration from environment: %w", err)
	}

	mdmConfig := &MDMConfig{
		MDMPrefix:       config.MDMPrefix,
		MDMToken:        config.MDMToken,
		SigningIdentity: config.SigningIdentity,
		CacheSize:       config.CacheSize,
		CacheTTL:        config.CacheTTL,
		CachePrefix:     config.CachePrefix,
		ProfileConfig: &ProfileConfig{
			PayloadVersion:      config.PayloadVersion,
			PayloadIdentifier:   config.PayloadIdentifier,
			PayloadUUID:         config.PayloadUUID,
			PayloadOrganization: config.PayloadOrganization,
		},
	}

	mdm, err := NewMDM(mdmConfig)
	if err != nil {
		return fmt.Errorf("could not create mdm: %w", err)
	}

	h := &HTTPService{MDM: mdm}

	r := mux.NewRouter()

	lmt := tollbooth.NewLimiter(config.DeliverRate/60, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour}).
		SetIPLookups([]string{"RemoteAddr"})
	r.Methods("POST").Path("/v1/lsrelay/deliver").Handler(
		tollbooth.LimitHandler(lmt,
			h.DeliverHandler()))

	lmt = tollbooth.NewLimiter(config.FileRate/60, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour}).
		SetIPLookups([]string{"RemoteAddr"})
	r.Methods("HEAD", "GET").PathPrefix("/v1/lsrelay/files/").Handler(
		http.StripPrefix("/v1/lsrelay/files/",
			tollbooth.LimitHandler(lmt,
				h.FileStoreHandler())))

	logger := NewLogger(os.Stdout)

	handler := LogHandler(logger, r)
	if config.ProxyHeaders {
		handler = handlers.ProxyHeaders(handler)
	}

	return http.ListenAndServe(config.ListenAddr, handler)
}

func main() {
	err := RunServer()
	if err != nil {
		fmt.Println("Error: could not start server:", err)
	}
}
