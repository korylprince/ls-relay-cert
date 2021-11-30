package main

import (
	"fmt"
	"net/http"
	"os"

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
			PayloadOrganization: config.PayloadUUID,
		},
	}

	mdm, err := NewMDM(mdmConfig)
	if err != nil {
		return fmt.Errorf("could not create mdm: %w", err)
	}

	h := &HTTPService{MDM: mdm}

	r := mux.NewRouter()
	r.Methods("POST").Path("/v1/lsrelay/deliver").Handler(h.DeliverHandler())
	r.Methods("HEAD", "GET").PathPrefix("/v1/lsrelay/files/").Handler(http.StripPrefix("/v1/lsrelay/files/", h.FileStoreHandler()))

	logger := NewLogger(os.Stdout)

	return http.ListenAndServe(config.ListenAddr, LogHandler(logger, r))
}

func main() {
	err := RunServer()
	if err != nil {
		fmt.Println("Error: could not start server:", err)
	}
}
