package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/didip/tollbooth"
	"github.com/didip/tollbooth/limiter"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/korylprince/ls-relay-cert/mdm"
	"github.com/korylprince/ls-relay-cert/profile"
)

// LimitHandler is a middleware that performs rate-limiting given http.Handler struct.
func LimitHandler(lmt *limiter.Limiter, next http.Handler) http.Handler {
	type response struct {
		Code        int    `json:"code"`
		Description string `json:"description"`
	}

	middle := func(w http.ResponseWriter, r *http.Request) {
		httpError := tollbooth.LimitByRequest(lmt, w, r)
		if httpError != nil {
			lmt.ExecOnLimitReached(w, r)

			body := response{Code: httpError.StatusCode, Description: http.StatusText(httpError.StatusCode)}
			l := r.Context().Value(ContextKeyLog).(*Log)
			l.Error = httpError.Error()

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(httpError.StatusCode)

			e := json.NewEncoder(w)
			err := e.Encode(body)
			if err != nil {
				l.Error = err.Error()
			}
			return
		}

		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(middle)
}

func RunServer() error {
	config := new(Config)
	err := envconfig.Process("", config)
	if err != nil {
		return fmt.Errorf("could not process configuration from environment: %w", err)
	}

	mdmConfig := &mdm.Config{
		MDMPrefix:       config.MDMPrefix,
		MDMToken:        config.MDMToken,
		SigningIdentity: config.SigningIdentity,
		CacheSize:       config.CacheSize,
		CacheTTL:        config.CacheTTL,
		CachePrefix:     config.CachePrefix,
		Config: &profile.Config{
			PayloadVersion:      config.PayloadVersion,
			PayloadIdentifier:   config.PayloadIdentifier,
			PayloadUUID:         config.PayloadUUID,
			PayloadOrganization: config.PayloadOrganization,
		},
	}

	mdm, err := mdm.New(mdmConfig)
	if err != nil {
		return fmt.Errorf("could not create mdm: %w", err)
	}

	h := &HTTPService{MDM: mdm}

	r := mux.NewRouter()

	lmt := limiter.New(&limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour}).
		SetMax(float64(config.DeliverRate) / 60).
		SetBurst(config.DeliverRate).
		SetIPLookups([]string{"RemoteAddr"})
	r.Methods("POST").Path("/v1/lsrelay/deliver").Handler(
		LimitHandler(lmt,
			h.DeliverHandler()))

	lmt = limiter.New(&limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour}).
		SetMax(float64(config.FileRate) / 60).
		SetBurst(config.FileRate).
		SetIPLookups([]string{"RemoteAddr"})
	r.Methods("HEAD", "GET").PathPrefix("/v1/lsrelay/files/").Handler(
		http.StripPrefix("/v1/lsrelay/files/",
			LimitHandler(lmt,
				h.FileStoreHandler())))

	logger := NewLogger(os.Stdout)

	handler := LogHandler(logger, r)
	if config.ProxyHeaders {
		handler = handlers.ProxyHeaders(handler)
	}

	fmt.Println("Listening on:", config.ListenAddr)

	return http.ListenAndServe(config.ListenAddr, handler)
}

func main() {
	err := RunServer()
	if err != nil {
		fmt.Println("Error: could not start server:", err)
	}
}
