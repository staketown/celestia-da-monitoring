package main

import (
	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/go-square/v2/share"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	namespaces = []InternalNamespace{
		{
			Name:      "TxNamespace",
			Namespace: share.TxNamespace,
		},
		{
			Name:      "IntermediateStateRootsNamespace",
			Namespace: share.IntermediateStateRootsNamespace,
		},
		{
			Name:      "PayForBlobNamespace",
			Namespace: share.PayForBlobNamespace,
		},
		{
			Name:      "PrimaryReservedPaddingNamespace",
			Namespace: share.PrimaryReservedPaddingNamespace,
		},
		{
			Name:      "MaxPrimaryReservedNamespace",
			Namespace: share.MaxPrimaryReservedNamespace,
		},
		{
			Name:      "MinSecondaryReservedNamespace",
			Namespace: share.MinSecondaryReservedNamespace,
		},
	}
)

type InternalNamespace struct {
	Name      string
	Namespace share.Namespace
}

func SharesHandler(w http.ResponseWriter, r *http.Request, rpcClient *client.Client) {
	requestStart := time.Now()

	sublogger := log.With().
		Str("request-id", uuid.New().String()).
		Logger()

	sharesByNamespaceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name:        "shares_by_namespace",
			Help:        "All shares with proofs within a specific namespace by local height",
			ConstLabels: ConstLabels,
		},
		[]string{
			"namespace",
			"local_height",
		},
	)

	registry := prometheus.NewRegistry()
	registry.MustRegister(sharesByNamespaceGauge)

	sublogger.Debug().Msg("Started querying shares data")
	peersDataStart := time.Now()

	localHead, err := rpcClient.Header.LocalHead(ctx)

	if err != nil {
		sublogger.Error().
			Err(err).
			Msg("Could not get local head data")
		return
	}

	var wg sync.WaitGroup

	for _, ns := range namespaces {
		wg.Add(1)

		go func(ns InternalNamespace) {
			defer wg.Done()

			sharesByNamespaceResponse, err := rpcClient.Share.GetNamespaceData(ctx, localHead.Height(), ns.Namespace)

			if err != nil {
				sublogger.Error().
					Err(err).
					Msgf("Could not get shares by namespace: %s", ns)
				return
			}

			sharesByNamespaceGauge.With(prometheus.Labels{
				"namespace":    ns.Name,
				"local_height": strconv.FormatUint(localHead.Height(), 10),
			}).Set(float64(len(sharesByNamespaceResponse.Flatten())))

		}(ns)
	}

	wg.Wait()

	sublogger.Debug().
		Float64("request-time", time.Since(peersDataStart).Seconds()).
		Msg("Finished querying shares data")

	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	h.ServeHTTP(w, r)
	sublogger.Info().
		Str("method", "GET").
		Str("endpoint", "/metrics/shares").
		Float64("request-time", time.Since(requestStart).Seconds()).
		Msg("Request processed")
}
