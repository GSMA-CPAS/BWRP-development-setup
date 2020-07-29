#!/bin/bash

. cli/utils.sh

function setup {
  createChannel dtag peer0 7051
  joinChannel dtag peer0 7051
  updateAnchorPeer dtag peer0 7051
  installPackagedChaincode dtag peer0 7051 $CHAINCODE_VERSION
  approveChaincode dtag peer0 7051 $CHAINCODE_VERSION 1
}

function invoke {
  invokeChaincode $1 $2 7051 '["createOrUpdateOrganization","{\"companyName\":\"Telekom\",\"storageEndpoint\":\"https\"}"]'
  #invokeChaincode $1 $2 7051 '["storeSignature","24","304402207f694c7075058ef8b01a98e62563a6b5c01243fe59a9ff9095c874db3fb8916502203dd6eb451f0990e33e1e079b1ae7aa5baa59b7dedbd2ca4ed928cffe8f497e3c","ecdsa-with-SHA256_secp256r1"]'
}

function query {
  queryChaincode $1 $2 7051 '["getOrganization"]'
  #queryChaincode $1 $2 7051 '["getSignatures","24"]'
}

function package_chaincode {
  packageChaincode $1 $2 7051 $3 $4
}

function install_chaincode {
  installPackagedChaincode $1 $2 7051 $3
}

function approve_chaincode {
  approveChaincode $1 $2 7051 $3 $4
}

function commit_chaincode {
  commitChaincode $1 $2 7051 $3 $4
}

function upgrade_chaincode {
  installPackagedChaincode dtag peer0 7051 $1
  installPackagedChaincode dtag peer1 8051 $1
  approveChaincode dtag peer0 7051 $1 $2
}

function update_crl {
  ORG=${1^^}
  fetchChannelConfig roaming-contracts
  createConfigUpdateWithCRL $ORG
}

function test {
  echo "test"
}

MODE=$1;shift

if [ "${MODE}" == "setup" ]; then
  setup
elif [ "${MODE}" == "invoke" ]; then
  invoke $1 $2
elif [ "${MODE}" == "query" ]; then
  query $1 $2
elif [ "${MODE}" == "package-chaincode" ]; then
  package_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "install-chaincode" ]; then
  install_chaincode $1 $2 $3
elif [ "${MODE}" == "approve-chaincode" ]; then
  approve_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "commit-chaincode" ]; then
  commit_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "upgrade-chaincode" ]; then
  upgrade_chaincode $1 $2
elif [ "${MODE}" == "info" ]; then
  channelInfo $1 $2 7051
elif [ "${MODE}" == "update-crl" ]; then
  update_crl $1
elif [ "${MODE}" == "test" ]; then
  test
else
  echo "Usage: "
  echo "   script.sh setup|invoke|query|package-chaincode|install-chaincode|approve-chaincode|commit-chaincode|info"
  exit 1
fi
