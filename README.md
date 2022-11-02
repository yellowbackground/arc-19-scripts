# arc-19-scripts

Script to patch arc-19/arc-3 assets.

!!! I recommended running this script against your assets on testnet before running against mainnet. It was only tested against FUMs by Hans which inspired this script. 

## Limitations

* Currently only supports arc3 traits. It should be fairly simple to support other use cases and PRs are encouraged.

## Prerequisites

* Your collection is already minted - this script only patches existing ASAs
* The script depends on Golang being installed on your machine. You can install from here: https://go.dev/dl/
* Nft Storage is used to publish files to IPFS. You'll need to generate an API key and add it to `config.json` 

## Running the script

1. Add assets and json metadata files to the `assets` directory. See examples
2. Create a file named `config.json` in the root. You can copy `config.example.json`.
3. Run the script `go run main.go`.

## Config

```json
{
  "algodUrl": "https://mainnet-api.algonode.cloud",
  "mnemonic": "25 word passhphrase here",
  "nftStorageApiKey": "your nft storage api key here",
  "namePrefix": "FUMS #",
  "imageExtension": ".jpeg",
  "imageMimeType": "image/jpeg",
  "description": "Welcome To The Fumly - 4000 1/1 NFTs",
  "assets": [
    {
      "index": 120051643,
      "number": "0881"
    }
  ]
}
```