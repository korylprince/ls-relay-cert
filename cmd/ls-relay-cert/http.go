package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/korylprince/ls-relay-cert/mdm"
)

type HTTPService struct {
	*mdm.MDM
}

// DeliverHandler delivers the payload to the serial number specified in the request
func (s *HTTPService) DeliverHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := r.Context().Value(ContextKeyLog).(*Log)

		code, body := func(w http.ResponseWriter, r *http.Request) (int, interface{}) {
			type request struct {
				SerialNumber string `json:"serial_number"`
			}

			req := new(request)
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(req); err != nil {
				return http.StatusBadRequest, fmt.Errorf("could not parse request: %w", err)
			}

			if req.SerialNumber == "" {
				return http.StatusBadRequest, errors.New("empty serial_number")
			}

			l.SerialNumber = req.SerialNumber

			if err := s.Deliver(req.SerialNumber); err != nil {
				if errors.Is(err, mdm.ErrNotFound) {
					return http.StatusNotFound, err
				}
				return http.StatusInternalServerError, fmt.Errorf("could not deliver payload: %w", err)
			}

			return http.StatusOK, nil
		}(w, r)

		type response struct {
			Code        int    `json:"code"`
			Description string `json:"description"`
		}

		if err, ok := body.(error); ok || body == nil {
			resp := response{Code: code, Description: http.StatusText(code)}
			body = resp
			if ok {
				l.Error = err.Error()
			}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)

		e := json.NewEncoder(w)
		err := e.Encode(body)

		if err != nil {
			l.Error = err.Error()
		}
	})
}

// FileStoreHandler is a file handler. If the handler is not mounted at "/", then it should be wrapped in http.StripPrefix so the handler sees the request rooted at /
func (s *HTTPService) FileStoreHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		l := r.Context().Value(ContextKeyLog).(*Log)
		path := r.URL.Path
		var (
			file []byte
			err  error
		)
		if r.Method == http.MethodHead {
			file, err = s.Peek(path)
		} else {
			file, err = s.Get(path)
		}

		if err != nil {
			l.Error = err.Error()

			if errors.Is(err, mdm.ErrNotFound) {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("404 Not Found"))
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("500 Internal Server Error"))
			return
		}

		http.ServeContent(w, r, "payload.pkg", time.Now(), bytes.NewReader(file))
	})
}
