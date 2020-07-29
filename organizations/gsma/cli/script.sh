#!/bin/bash

. cli/utils.sh

function setup {
  fetchChannel gsma peer0 10051
  joinChannel gsma peer0 10051
  updateAnchorPeer gsma peer0 10051
  installPackagedChaincode gsma peer0 10051 $CHAINCODE_VERSION
  approveChaincode gsma peer0 10051 $CHAINCODE_VERSION 1
  commitChaincode gsma peer0 10051 $CHAINCODE_VERSION 1
}

function invoke {
  invokeChaincode $1 $2 10051 '["createOrUpdateOrganization","{\"companyName\":\"GSMA\",\"storageEndpoint\":\"https\"}"]'
}

function query {
  queryChaincode $1 $2 10051 '["getOrganization"]'
}

function package_chaincode {
  packageChaincode $1 $2 10051 $3 $4
}

function install_chaincode {
  installPackagedChaincode $1 $2 10051 $3
}

function approve_chaincode {
  approveChaincode $1 $2 10051 $3 $4
}

function commit_chaincode {
  commitChaincode $1 $2 10051 $3 $4
}

function upgrade_chaincode {
  installPackagedChaincode gsma peer0 10051 $1
  approveChaincode gsma peer0 10051 $1 $2
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
  channelInfo $1 $2 10051
elif [ "${MODE}" == "update-crl" ]; then
  update_crl $1
elif [ "${MODE}" == "test" ]; then
  test
else
  echo "Usage: "
  echo "   script.sh setup|invoke|query|package-chaincode|install-chaincode|approve-chaincode|commit-chaincode|info"
  exit 1
fi
