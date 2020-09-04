#!/bin/bash

ORDERER="orderer.nomad.com:7050"
ORDERER_CA=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nomad.com/orderers/orderer.nomad.com/msp/tlscacerts/tlsca.nomad.com-cert.pem
ORDERER_CLIENT_CERTFILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nomad.com/users/Admin@nomad.com/tls/client.crt
ORDERER_CLIENT_KEYFILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/ordererOrganizations/nomad.com/users/Admin@nomad.com/tls/client.key
CHANNEL_NAME=roaming-contracts

CHAINCODE_NAME="hybrid"
CHAINCODE_VERSION_DEFAULT="0.0.1"

#PEER_CONN_PARMS="--peerAddresses peer0.dtag.nomad.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dtag.nomad.com/peers/peer0.dtag.nomad.com/tls/ca.crt"
PEER_CONN_PARMS="--peerAddresses peer0.dtag.nomad.com:7051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/dtag.nomad.com/peers/peer0.dtag.nomad.com/tls/ca.crt --peerAddresses peer0.tmus.nomad.com:9051 --tlsRootCertFiles /opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/tmus.nomad.com/peers/peer0.tmus.nomad.com/tls/ca.crt"


setGlobals() {
  export CORE_PEER_LOCALMSPID=${ORG^^}
  export CORE_PEER_MSPCONFIGPATH=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/${ORG}.nomad.com/users/Admin@${ORG}.nomad.com/msp
  export CORE_PEER_ADDRESS=${PEER}.${ORG}.nomad.com:${PORT}
  export CORE_PEER_TLS_CERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/${ORG}.nomad.com/peers/${PEER}.${ORG}.nomad.com/tls/server.crt
  export CORE_PEER_TLS_KEY_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/${ORG}.nomad.com/peers/${PEER}.${ORG}.nomad.com/tls/server.key
  export CORE_PEER_TLS_ROOTCERT_FILE=/opt/gopath/src/github.com/hyperledger/fabric/peer/crypto/peerOrganizations/${ORG}.nomad.com/peers/${PEER}.${ORG}.nomad.com/tls/ca.crt
}
#always set globals
setGlobals

createChannel() {
  set -x
	peer channel create -o $ORDERER -c ${CHANNEL_NAME} -f ./channel-artifacts/channel-${CHANNEL_NAME}.tx --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
	set +x
}

fetchChannel() {
  set -x
	peer channel fetch 0 -o $ORDERER -c ${CHANNEL_NAME} ${CHANNEL_NAME}.block --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
	set +x
}

joinChannel() {
	set -x
	peer channel join -b ${CHANNEL_NAME}.block
  res=$?
	set +x
}

# Update channel configuration to define anchor peer
updateAnchorPeer() {
	set -x
	peer channel update -o $ORDERER -c ${CHANNEL_NAME} -f ./channel-artifacts/channel-${CHANNEL_NAME}-anchor-${ORG}.tx --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
	set +x
}

packageChaincode() {
  set -x
  CHAINCODE_VERSION=${1:-$CHAINCODE_VERSION_DEFAULT}
  peer lifecycle chaincode package cli/${CHAINCODE_NAME}-v${CHAINCODE_VERSION}.tar.gz --path /opt/gopath/src/github.com/chaincode/${CHAINCODE_NAME}/ --label "${CHAINCODE_NAME}_v${CHAINCODE_VERSION}"
  res=$?
  set +x
}

installPackagedChaincode() {
  set -x
  CHAINCODE_VERSION=${1:-$CHAINCODE_VERSION_DEFAULT}
  peer lifecycle chaincode install cli/${CHAINCODE_NAME}-v${CHAINCODE_VERSION}.tar.gz
  res=$?
  set +x
}

approveChaincode() {
  SEQUENCE=$1
  CHAINCODE_VERSION=${2:-$CHAINCODE_VERSION_DEFAULT}
  set -x
  peer lifecycle chaincode queryinstalled >&log.txt
  res=$?
  set +x
  cat log.txt
  PACKAGE_ID=$(sed -n "/${CHAINCODE_NAME}_v${VERSION}/{s/^Package ID: //; s/, Label:.*$//; p;}" log.txt)
  echo PackageID is ${PACKAGE_ID}
  set -x
  peer lifecycle chaincode approveformyorg -o $ORDERER -C ${CHANNEL_NAME} -n $CHAINCODE_NAME -v $CHAINCODE_VERSION --package-id $PACKAGE_ID --sequence $SEQUENCE --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
  set +x
}

commitChaincode() {
  SEQUENCE=$1
  CHAINCODE_VERSION=${2:-$CHAINCODE_VERSION_DEFAULT}
  set -x
  peer lifecycle chaincode checkcommitreadiness -C ${CHANNEL_NAME} -n $CHAINCODE_NAME -v $CHAINCODE_VERSION --sequence $SEQUENCE --output json --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
  set +x
  echo '-- commit --'
  set -x
  peer lifecycle chaincode commit -o $ORDERER -C ${CHANNEL_NAME} -n $CHAINCODE_NAME -v $CHAINCODE_VERSION --sequence $SEQUENCE $PEER_CONN_PARMS --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
  set +x
}

queryChaincode() {
  FUNC=$1
  set -x
	peer chaincode query -C ${CHANNEL_NAME} -n $CHAINCODE_NAME -c '{"Args":'${FUNC}'}'
  res=$?
  set +x
}

invokeChaincode() {
  FUNC=$1
  set -x
  peer chaincode invoke -o $ORDERER -C ${CHANNEL_NAME} -n $CHAINCODE_NAME -c '{"Args":'${FUNC}'}' $PEER_CONN_PARMS --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
  set +x
}

invokeWithTransientChaincode() {
  FUNC=$1
  TRANSIENT=$2
  set -x
  peer chaincode invoke -o $ORDERER -C ${CHANNEL_NAME} -n $CHAINCODE_NAME -c '{"Args":'${FUNC}'}' --transient $TRANSIENT $PEER_CONN_PARMS --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  res=$?
  set +x
}

channelInfo() {
  peer channel list
  peer lifecycle chaincode queryinstalled
  peer lifecycle chaincode querycommitted -C ${CHANNEL_NAME} --output json
}

fetchChannelConfig() {
  CHANNEL_ID=$1
  set -x
  # fetch most recent config block for the channel
	# peer channel fetch config config_block.pb -o $ORDERER -c $CHANNEL_ID --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  peer channel fetch config config_block.pb -o $ORDERER -c $CHANNEL_ID --tls --cafile $ORDERER_CA
  # decode channel configuration block into JSON format
  # strip away all of the headers, metadata, creator signatures, and so on that are irrelevant to the change we want to make
  configtxlator proto_decode --input config_block.pb --type common.Block | jq .data.data[0].payload.data.config > config.json
  res=$?
	set +x
}

# createConfigUpdate <channel_id> <original_config.json> <modified_config.json> <output.pb>
createConfigUpdate() {
  CHANNEL_ID=$1
  ORIGINAL=$2
  MODIFIED=$3
  OUTPUT=$4

  set -x
  # translate config.json back to protobuf
  configtxlator proto_encode --input "${ORIGINAL}" --type common.Config --output original_config.pb
  # encode modified_config.json to modified_config.pb
  configtxlator proto_encode --input "${MODIFIED}" --type common.Config --output modified_config.pb
  # calculate delta between these two config protobufs
  configtxlator compute_update --channel_id "${CHANNEL_ID}" --original original_config.pb --updated modified_config.pb --output config_update.pb
  # decode config_update.pd into JSON format
  configtxlator proto_decode --input config_update.pb  --type common.ConfigUpdate --output config_update.json
  # Now, we have a decoded update file – config_update.json – that we need to wrap in an envelope message. This step will give us back the header field that we stripped away earlier
  echo '{"payload":{"header":{"channel_header":{"channel_id":"'$CHANNEL_ID'", "type":2}},"data":{"config_update":'$(cat config_update.json)'}}}' | jq . > config_update_in_envelope.json
  # encode config_update_in_envelope.json to protobuf format
  configtxlator proto_encode --input config_update_in_envelope.json --type common.Envelope --output "${OUTPUT}"
  set +x
}

# Steps before:
# 1. Create Certificate Revocation List (crl.pem)
# 2. Copy crl.pem to $CORE_PEER_MSPCONFIGPATH/crls/crl.pem
createConfigUpdateWithCRL() {
  crl=$(cat $CORE_PEER_MSPCONFIGPATH/crls/crl*.pem | base64 | tr -d '\n')
  cat config.json | jq '.channel_group.groups.Application.groups.'"${ORG}"'.values.MSP.value.config.revocation_list = ["'"${crl}"'"]' > updated_config.json
  #cat config.json | jq '.channel_group.groups.Application.groups.'"${ORG}"'.values.MSP.value.config.revocation_list = []' > updated_config.json
  createConfigUpdate ${CHANNEL_NAME} config.json updated_config.json output.pb
  # peer channel update -f output.pb -c ${CHANNEL_NAME} -o $ORDERER --tls --cafile $ORDERER_CA --clientauth --certfile $ORDERER_CLIENT_CERTFILE --keyfile $ORDERER_CLIENT_KEYFILE
  peer channel update -f output.pb -c ${CHANNEL_NAME} -o $ORDERER --tls --cafile $ORDERER_CA
}

signConfig() {
  FILE=$1
  set -x
  peer channel signconfigtx -f $FILE
  res=$?
  set +x
}

function setupChannel {
  fetchChannel
  joinChannel
  updateAnchorPeer
}

function setupChaincode {
  SEQUENCE=$1
  #packageChaincode
  installPackagedChaincode
  approveChaincode $SEQUENCE
}

# Check if the function exists (bash specific)
if declare -f "$1" > /dev/null
then
  # call arguments verbatim
  "$@"
else
  # Show a helpful error
  echo "'$1' is not a known function name" >&2
  exit 1
fi
