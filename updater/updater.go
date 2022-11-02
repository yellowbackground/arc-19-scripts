package updater

import (
	"context"
	"crypto/ed25519"
	"encoding/json"
	"fmt"
	"github.com/algorand/go-algorand-sdk/client/v2/algod"
	"github.com/algorand/go-algorand-sdk/crypto"
	"github.com/algorand/go-algorand-sdk/future"
	"github.com/algorand/go-algorand-sdk/mnemonic"
	"github.com/algorand/go-algorand-sdk/types"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/yellowbackground/arc-19-scripts/config"
	"github.com/yellowbackground/arc-19-scripts/nftstorage"
	"os"
)

type arc3Metadata struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Image       string            `json:"image"`
	Properties  map[string]string `json:"properties"`
}

func UpdateAssets(conf config.Config, algodClient *algod.Client, client nftstorage.Client) error {
	for _, asset := range conf.Assets {
		log.Info().Msgf("Updating asset #%s", asset.Number)
		if err := updateAsset(conf, algodClient, client, asset); err != nil {
			return errors.Wrapf(err, "failed to update asset #%s", asset.Number)
		}
	}
	return nil
}

func updateAsset(conf config.Config, algodClient *algod.Client, nftStorageClient nftstorage.Client, asset config.Asset) error {
	imageCID, err := uploadImageToIPFS(conf, nftStorageClient, asset)
	if err != nil {
		return err
	}

	arc3MetadataCID, err := uploadArc3MetadataToIPFS(conf, nftStorageClient, asset, imageCID)
	if err != nil {
		return err
	}

	reserveAddress, err := getReserveAddressFromCID(arc3MetadataCID)
	if err != nil {
		return err
	}

	return executeAssetConfigTransaction(conf, algodClient, asset, reserveAddress)
}

func executeAssetConfigTransaction(conf config.Config, algodClient *algod.Client, asset config.Asset, reserveAddress string) error {
	suggestedParams, err := algodClient.SuggestedParams().Do(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to get suggested params")
	}

	privateKey, err := derivePrivateKeyFromMnemonic(conf.Mnemonic)
	if err != nil {
		return errors.Wrap(err, "failed to get private key from mnemonic")
	}

	creatorAddress, err := derivePublicAddressFromMnemonic(conf.Mnemonic)
	if err != nil {
		return errors.Wrap(err, "failed to get public key from mnemonic")
	}

	txn, err := future.MakeAssetConfigTxn(
		creatorAddress,
		nil,
		suggestedParams,
		asset.Index,
		creatorAddress,
		reserveAddress,
		"",
		"",
		false,
	)
	if err != nil {
		return errors.Wrap(err, "failed to create asset config tx")
	}

	txID, signedTxn, err := crypto.SignTransaction(privateKey, txn)
	if err != nil {
		return errors.Wrap(err, "failed to sign asset config tx")
	}

	_, err = algodClient.SendRawTransaction(signedTxn).Do(context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to submit asset config tx")
	}

	_, err = future.WaitForConfirmation(algodClient, txID, 4, context.Background())
	if err != nil {
		return errors.Wrap(err, "failed to wait for asset config tx confirmation")
	}

	return nil
}

func uploadImageToIPFS(conf config.Config, nftStorageClient nftstorage.Client, asset config.Asset) (string, error) {
	img, err := loadImage(conf, asset)
	if err != nil {
		return "", err
	}

	ipfsCID, err := nftStorageClient.UploadFile(img, conf.ImageMimeType)
	if err != nil {
		return "", errors.Wrap(err, "failed to upload image to Nft Storage")
	}

	return ipfsCID, nil
}

func uploadArc3MetadataToIPFS(conf config.Config, nftStorageClient nftstorage.Client, asset config.Asset, imageCID string) (string, error) {
	traits, err := loadTraits(asset)
	if err != nil {
		return "", err
	}

	arc3Json, err := renderArc3Json(conf, asset, traits, imageCID)
	if err != nil {
		return "", errors.Wrap(err, "failed to render arc3 json")
	}

	ipfsCID, err := nftStorageClient.UploadFile(arc3Json, "application/json")
	if err != nil {
		return "", errors.Wrap(err, "failed to upload arc3 json to Nft Storage")
	}

	return ipfsCID, nil
}

func loadTraits(asset config.Asset) (map[string]string, error) {
	fileName := fmt.Sprintf("assets/%s.json", asset.Number)
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read json file: %s", fileName)
	}
	var traits map[string]string
	if err := json.Unmarshal(file, &traits); err != nil {
		return nil, errors.Wrapf(err, "failed to decode json file: %s", fileName)
	}
	return traits, nil
}

func loadImage(conf config.Config, asset config.Asset) ([]byte, error) {
	fileName := fmt.Sprintf("assets/%s%s", asset.Number, conf.ImageExtension)
	file, err := os.ReadFile(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to read image file: %s", fileName)
	}
	return file, nil
}

func renderArc3Json(conf config.Config, asset config.Asset, traits map[string]string, imageCID string) ([]byte, error) {
	return json.Marshal(arc3Metadata{
		Name:        fmt.Sprintf("%s%s", conf.NamePrefix, asset.Number),
		Description: conf.Description,
		Image:       fmt.Sprintf("ipfs://%s", imageCID),
		Properties:  traits,
	})
}

func getReserveAddressFromCID(ipfsCID string) (string, error) {
	decodedCID, err := cid.Decode(ipfsCID)
	if err != nil {
		return "", errors.Wrap(err, "failed to decode cid")
	}

	decodedMultiHash, err := multihash.Decode(decodedCID.Hash())
	if err != nil {
		return "", errors.Wrap(err, "failed to decode ipfs cid")
	}

	reserve, err := types.EncodeAddress(decodedMultiHash.Digest)
	if err != nil {
		return "", errors.Wrap(err, "failed to generate reserve from cid")
	}

	return reserve, nil
}

func derivePrivateKeyFromMnemonic(passphrase string) (ed25519.PrivateKey, error) {
	var err error
	sk, err := mnemonic.ToPrivateKey(passphrase)
	if err != nil {
		return nil, fmt.Errorf("error with private key conversion. %w", err)
	}
	return sk, nil
}

func derivePublicAddressFromMnemonic(passphrase string) (string, error) {
	privateKey, err := derivePrivateKeyFromMnemonic(passphrase)
	if err != nil {
		return "", fmt.Errorf("error with private key conversion. %w", err)
	}
	pk := privateKey.Public()
	var a types.Address
	cpk := pk.(ed25519.PublicKey)
	copy(a[:], cpk[:])
	return a.String(), nil
}
