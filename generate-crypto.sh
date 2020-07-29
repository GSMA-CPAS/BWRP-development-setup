#!/bin/bash

# generate crypto config
#cryptogen generate --config=./crypto-config.yaml

# generate genesis block
configtxgen -profile NomadOrdererGenesis -outputBlock ./channel-artifacts/genesis.block -channelID system-channel

# generate channel creation transaction
configtxgen -profile RoamingContractsChannel -outputCreateChannelTx ./channel-artifacts/channel-roaming-contracts.tx -channelID roaming-contracts

# generate anchor peer transaction (DTAG)
configtxgen -profile RoamingContractsChannel -outputAnchorPeersUpdate ./channel-artifacts/channel-roaming-contracts-anchor-dtag.tx -channelID roaming-contracts -asOrg DTAG

# generate anchor peer transaction (TMUS)
configtxgen -profile RoamingContractsChannel -outputAnchorPeersUpdate ./channel-artifacts/channel-roaming-contracts-anchor-tmus.tx -channelID roaming-contracts -asOrg TMUS

# generate anchor peer transaction (GSMA)
configtxgen -profile RoamingContractsChannel -outputAnchorPeersUpdate ./channel-artifacts/channel-roaming-contracts-anchor-gsma.tx -channelID roaming-contracts -asOrg GSMA
