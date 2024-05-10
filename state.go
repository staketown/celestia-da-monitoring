package main

import (
	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func StateHandler(w http.ResponseWriter, r *http.Request, rpcClient *client.Client) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	addressBalanceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "address_balance",
			Help:        "Balance of the given address",
			ConstLabels: ConstLabels,
		},
		[]string{
			"address",
			"denom",
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(addressBalanceGauge)

	sublogger.Debug().Msg("Started querying balance for address data")
	peersDataStart := time.Now()

	balanceResponse, err := rpcClient.State.Balance(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get balance data")
		return
	}

	addressResponse, err := rpcClient.State.AccountAddress(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get address data")
		return
	}

	sublogger.Debug().
		Float64("request-time", time.Since(peersDataStart).Seconds()).
		Msg("Finished querying header data")

	addressBalanceGauge.With(prometheus.Labels{
		"address": addressResponse.String(),
		"denom":   balanceResponse.Denom,
	}).Set(float64(balanceResponse.Amount.Int64()))

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/state").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
