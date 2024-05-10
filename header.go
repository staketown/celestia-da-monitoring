package main

import (
	"fmt"
	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func HeaderHandler(w http.ResponseWriter, r *http.Request, rpcClient *client.Client) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	localHeadGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "local_height",
			Help:        "Local height of bridge node",
			ConstLabels: ConstLabels,
		},
	)

	localHeadTimeGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "local_height_time",
			Help:        "Local height of bridge node",
			ConstLabels: ConstLabels,
		},
	)

	networkHeadGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "network_height",
			Help:        "Consensus height",
			ConstLabels: ConstLabels,
		},
	)

	networkHeadTimeGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "network_height_time",
			Help:        "Consensus height",
			ConstLabels: ConstLabels,
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(localHeadGauge)
	registry.MustRegister(localHeadTimeGauge)
	registry.MustRegister(networkHeadGauge)
	registry.MustRegister(networkHeadTimeGauge)

	sublogger.Debug().Msg("Started querying header data")
	peersDataStart := time.Now()

	localHead, err := rpcClient.Header.LocalHead(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get local head data")
		return
	}

	networkHead, err := rpcClient.Header.NetworkHead(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get network head data")
		return
	}

	sublogger.Debug().
		Float64("request-time", time.Since(peersDataStart).Seconds()).
		Msg("Finished querying header data")

	localHeadGauge.Set(float64(localHead.Height()))
	localHeadTimeGauge.Set(float64(localHead.Time().Unix()))
	networkHeadGauge.Set(float64(networkHead.Height()))
	networkHeadTimeGauge.Set(float64(networkHead.Time().Unix()))

	response, _ := rpcClient.Share.GetSharesByNamespace(ctx, localHead, share.PayForBlobNamespace)
	fmt.Println(localHead.Height())
	fmt.Println(len(response.Flatten()))

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/header").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
