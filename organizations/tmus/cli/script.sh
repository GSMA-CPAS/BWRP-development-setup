#!/bin/bash

. cli/utils.sh

function setup {
  fetchChannel tmus peer0 9051
  joinChannel tmus peer0 9051
  updateAnchorPeer tmus peer0 9051
  installPackagedChaincode tmus peer0 9051 $CHAINCODE_VERSION
  approveChaincode tmus peer0 9051 $CHAINCODE_VERSION 1
}

function invoke {
  invokeChaincode $1 $2 9051 '["createOrUpdateOrganization","{\"companyName\":\"T-MobileUS\",\"storageEndpoint\":\"https\"}"]'
}

function query {
  queryChaincode $1 $2 9051 '["getOrganization"]'
}

function package_chaincode {
  packageChaincode $1 $2 9051 $3 $4
}

function install_chaincode {
  installPackagedChaincode $1 $2 9051 $3
}

function approve_chaincode {
  approveChaincode $1 $2 9051 $3 $4
}

function commit_chaincode {
  commitChaincode $1 $2 9051 $3 $4
}

function upgrade_chaincode {
  installPackagedChaincode tmus peer0 9051 $1
  approveChaincode tmus peer0 9051 $1 $2
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
  channelInfo $1 $2
elif [ "${MODE}" == "test" ]; then
  test
else
  echo "Usage: "
  echo "   script.sh setup|invoke|query|package-chaincode|install-chaincode|approve-chaincode|commit-chaincode|info"
  exit 1
fi
