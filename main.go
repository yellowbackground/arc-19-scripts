package main

import (
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/yellowbackground/arc-19-scripts/config"
	"github.com/yellowbackground/arc-19-scripts/nftstorage"
	"github.com/yellowbackground/arc-19-scripts/updater"
	"os"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	conf, err := config.Load()
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to load config")
	}
	log.Info().Msgf("Found %d assets in config", len(conf.Assets))

	algodClient, err := algod.MakeClient(conf.AlgodURL, "")
	if err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to create algod client")
	}

	httpClient := retryablehttp.NewClient()
	httpClient.Logger = nil
	nftStorageClient := nftstorage.Client{
		APIKey:     conf.NftStorageApiKey,
		HTTPClient: httpClient.StandardClient(),
	}

	if err := updater.UpdateAssets(conf, algodClient, nftStorageClient); err != nil {
		log.Fatal().Stack().Err(err).Msg("failed to update assets")
	}

	log.Info().Msg("Done")
}
