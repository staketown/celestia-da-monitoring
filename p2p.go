package main

import (
	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func P2pHandler(w http.ResponseWriter, r *http.Request, rpcClient *client.Client) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	peersGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "peers",
			Help:        "Number of all peers connectd to",
			ConstLabels: ConstLabels,
		},
	)

	blockedPeersGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "blocked_peers",
			Help:        "Number of blocked peers",
			ConstLabels: ConstLabels,
		},
	)

	// The TotalIn and TotalOut fields record cumulative bytes sent / received.
	// The RateIn and RateOut fields record bytes sent / received per second.

	bandwidthTotalInGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "bandwidth_total_in",
			Help:        "cumulative bytes sent",
			ConstLabels: ConstLabels,
		},
	)

	bandwidthRateInGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "bandwidth_rate_in",
			Help:        "cumulative bytes sent per second",
			ConstLabels: ConstLabels,
		},
	)

	bandwidthTotalOutGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "bandwidth_total_out",
			Help:        "cumulative bytes received",
			ConstLabels: ConstLabels,
		},
	)

	bandwidthRateOutGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name:        "bandwidth_rate_out",
			Help:        "cumulative bytes received per second",
			ConstLabels: ConstLabels,
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(peersGauge)
	registry.MustRegister(blockedPeersGauge)
	registry.MustRegister(bandwidthTotalInGauge)
	registry.MustRegister(bandwidthTotalOutGauge)
	registry.MustRegister(bandwidthRateInGauge)
	registry.MustRegister(bandwidthRateOutGauge)

	sublogger.Debug().Msg("Started querying p2p data")
	peersDataStart := time.Now()

	peers, err := rpcClient.P2P.Peers(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get peers list data")
		return
	}

	blockedPeers, err := rpcClient.P2P.ListBlockedPeers(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get blocked peers data")
		return
	}

	bandwidthStats, err := rpcClient.P2P.BandwidthStats(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get bandwidth data")
		return
	}

	//bandwidthStats.
	sublogger.Debug().
		Float64("request-time", time.Since(peersDataStart).Seconds()).
		Msg("Finished querying p2p data")

	blockedPeersGauge.Set(float64(len(blockedPeers)))
	peersGauge.Set(float64(len(peers)))
	bandwidthTotalInGauge.Set(float64(bandwidthStats.TotalIn))
	bandwidthTotalOutGauge.Set(float64(bandwidthStats.TotalOut))
	bandwidthRateOutGauge.Set(bandwidthStats.RateOut)
	bandwidthRateInGauge.Set(bandwidthStats.RateIn)

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/p2p").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
