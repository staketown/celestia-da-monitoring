package main

import (
	"context"
	"fmt"
	"github.com/celestiaorg/celestia-node/api/rpc/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"net/http"
	"os"
)

var (
	ListenAddress string
	NodeAddress   string

	LogLevel string

	Token string

	ConstLabels map[string]string
)

var log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).With().Timestamp().Logger()

var rootCmd = &cobra.Command{
	Use:  "oracle-exporter",
	Long: "Scrape the data about the validators set, specific validators or wallets in the Cosmos network.",
	Run:  Execute,
}

var ctx = context.Background()

func Execute(_ *cobra.Command, _ []string) {
	logLevel, err := zerolog.ParseLevel(LogLevel)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not parse log level")
	}

	zerolog.SetGlobalLevel(logLevel)

	log.Info().
		Str("--listen-address", ListenAddress).
		Str("--node", NodeAddress).
		Str("--log-level", LogLevel).
		Msg("Started with following parameters")

	config := sdk.GetConfig()
	config.Seal()

	if err != nil {
		log.Fatal().Err(err).Msg("Could not connect to RPC node")
	}

	rpcClient, err := client.NewClient(ctx, NodeAddress, Token)

	if err != nil {
		fmt.Print(err)
	}

	http.HandleFunc("/metrics/p2p", func(w http.ResponseWriter, r *http.Request) {
		P2pHandler(w, r, rpcClient)
	})

	http.HandleFunc("/metrics/header", func(w http.ResponseWriter, r *http.Request) {
		HeaderHandler(w, r, rpcClient)
	})

	http.HandleFunc("/metrics/state", func(w http.ResponseWriter, r *http.Request) {
		StateHandler(w, r, rpcClient)
	})

	http.HandleFunc("/metrics/shares", func(w http.ResponseWriter, r *http.Request) {
		SharesHandler(w, r, rpcClient)
	})

	log.Info().Str("address", ListenAddress).Msg("Listening")
	err = http.ListenAndServe(ListenAddress, nil)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}

func main() {
	rootCmd.PersistentFlags().StringVar(&ListenAddress, "listen-address", ":9300", "The address this exporter would listen on")
	rootCmd.PersistentFlags().StringVar(&NodeAddress, "node", "http://127.0.0.0:26658", "RPC bridge node address")
	rootCmd.PersistentFlags().StringVar(&Token, "token", "eyJhbGciOiJ.celestia.testing", "OAuth token to bridge node with admin rights")
	rootCmd.PersistentFlags().StringVar(&LogLevel, "log-level", "info", "Logging level")

	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("Could not start application")
	}
}
