package config

import (
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type Config struct {
	AlgodURL         string  `json:"algodUrl"`
	Mnemonic         string  `json:"mnemonic"`
	NamePrefix       string  `json:"namePrefix"`
	Description      string  `json:"description"`
	ImageExtension   string  `json:"imageExtension"`
	ImageMimeType    string  `json:"imageMimeType"`
	NftStorageApiKey string  `json:"nftStorageApiKey"`
	Assets           []Asset `json:"assets"`
}

type Asset struct {
	Index  uint64 `json:"index"`
	Number string `json:"number"`
}

func Load() (Config, error) {
	jsonFile, err := os.ReadFile("./config.json")
	if err != nil {
		return Config{}, errors.WithStack(err)
	}
	var cfg Config
	err = json.Unmarshal(jsonFile, &cfg)
	if err != nil {
		return Config{}, errors.WithStack(err)
	}
	return cfg, nil
}
