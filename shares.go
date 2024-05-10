package main

import (
	"github.com/celestiaorg/celestia-node/api/rpc/client"
	"github.com/celestiaorg/celestia-node/share"
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
			Name:      "PayForBlobNamespace",
			Namespace: share.PayForBlobNamespace,
		},
		{
			Name:      "TxNamespace",
			Namespace: share.TxNamespace,
		},
		{
			Name:      "ISRNamespace",
			Namespace: share.ISRNamespace,
		},
		{
			Name:      "MaxPrimaryReservedNamespace",
			Namespace: share.MaxPrimaryReservedNamespace,
		},
		{
			Name:      "PrimaryReservedPaddingNamespace",
			Namespace: share.PrimaryReservedPaddingNamespace,
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
			Help:        "Total amount of shares per namespace by local height",
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

	for _, namespace := range namespaces {
		wg.Add(1)

		go func(namespace InternalNamespace) {
			defer wg.Done()

			sharesByNamespaceResponse, err := rpcClient.Share.GetSharesByNamespace(ctx, localHead, namespace.Namespace)

			if err != nil {
				sublogger.Error().
					Err(err).
					Msgf("Could not get shares by namespace: %s", namespace)
				return
			}

			sharesByNamespaceGauge.With(prometheus.Labels{
				"namespace":    namespace.Name,
				"local_height": strconv.FormatUint(localHead.Height(), 10),
			}).Set(float64(len(sharesByNamespaceResponse.Flatten())))

		}(namespace)
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
