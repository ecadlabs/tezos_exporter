package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	tezos "github.com/ecadlabs/tezos_exporter/go-tezos"
	log "github.com/sirupsen/logrus"
)

type HealthHandler struct {
	service   *tezos.Service
	interval  time.Duration
	chainID   string
	threshold int
	tcount    int
	ok        bool
}

func (h *HealthHandler) poll() {
	status, err := h.service.GetBootstrapped(context.Background(), h.chainID)
	if err != nil {
		log.WithError(err).Error("error getting bootstrap status")
		h.ok = false
	} else {
		h.ok = status.Bootstrapped && status.SyncState == tezos.SyncStateSynced
	}
	h.tcount = h.threshold

	tick := time.Tick(h.interval)
	for range tick {
		status, err := h.service.GetBootstrapped(context.Background(), h.chainID)
		if err != nil {
			log.WithError(err).Error("error getting bootstrap status")
			h.ok = false
			h.tcount = h.threshold
			continue
		}

		ok := status.Bootstrapped && status.SyncState == tezos.SyncStateSynced
		if ok != h.ok {
			h.tcount--
			if h.tcount == 0 {
				h.tcount = h.threshold
				h.ok = ok
			}
		} else {
			h.tcount = h.threshold
		}
	}
}

func (h *HealthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var res struct {
		Bootstrapped bool `json:"bootstrapped"`
	}

	var status int
	if h.ok {
		status = http.StatusOK
		res.Bootstrapped = true
	} else {
		status = http.StatusInternalServerError
		res.Bootstrapped = false
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(&res)
}

func NewHealthHandler(service *tezos.Service, chainID string, interval time.Duration, threshold int) *HealthHandler {
	h := HealthHandler{
		service:   service,
		interval:  interval,
		threshold: threshold,
		chainID:   chainID,
	}
	go h.poll()
	return &h
}
