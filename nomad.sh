#!/bin/bash

function printHelp() {
  echo "Usage: "
  echo "  nomad.sh setup"
  echo "  nomad.sh setup-explorer"
  echo "  nomad.sh package-chaincode <org> <peer> <version> <version>"
  echo "  nomad.sh install-chaincode <org> <peer> <version>"
  echo "  nomad.sh approve-chaincode <org> <peer> <version> <sequence>"
  echo "  nomad.sh commit-chaincode <org> <peer> <version> <sequence>"
  echo "  nomad.sh upgrade-chaincode <org> <peer> <version> <sequence>"
  echo "  nomad.sh update-crl <org>"
  echo "  nomad.sh info <org> <peer>"
  echo "  nomad.sh tty <container-name>"
  echo "  nomad.sh rebuild <name>"
  echo "  nomad.sh up"
  echo "  nomad.sh down"
  echo
}

function setupChannel() {
  echo "setting up channel"
  docker exec -ti cli-gsma cli/utils.sh createChannel
  docker exec -ti cli-gsma cli/utils.sh setupChannel
  docker exec -ti cli-dtag cli/utils.sh setupChannel
  docker exec -ti cli-tmus cli/utils.sh setupChannel
}

function setupChaincodes() {
  echo "setting up chaincodes"
  docker exec -ti cli-gsma cli/utils.sh setupChaincode 1
  docker exec -ti cli-dtag cli/utils.sh setupChaincode 1
  docker exec -ti cli-tmus cli/utils.sh setupChaincode 1
  docker exec -ti cli-gsma cli/utils.sh commitChaincode 1
}

function setupWebapp() {
  docker-compose restart blockchain-adapter-dtag
  docker-compose restart blockchain-adapter-tmus
  curl -s -X PUT http://localhost:8081/config/offchain-db-adapter -d '{"restURI": "http://offchain-db-adapter-dtag:3333"}' -H "Content-Type: application/json" > /dev/null
  curl -s -X PUT http://localhost:8082/config/offchain-db-adapter -d '{"restURI": "http://offchain-db-adapter-tmus:3334"}' -H "Content-Type: application/json" > /dev/null
  echo "setting up webapp"
  docker exec -ti --user nomad webapp-dtag node setup.js
  docker exec -ti --user nomad webapp-tmus node setup.js
}

function upgradeChaincodes() {
  echo "upgrade chaincodes"
  CURRENT_VERSION=$(docker exec -ti cli-gsma peer lifecycle chaincode querycommitted -C roaming-contracts | tail -n1 | cut -f4 -d" " | cut -d"," -f1)
  CURRENT_SEQUENCE=$(docker exec -ti cli-gsma peer lifecycle chaincode querycommitted -C roaming-contracts | tail -n1 | cut -f6 -d" " | cut -d"," -f1)
  NEW_VERSION=$(echo ${CURRENT_VERSION} | awk -F. -v OFS=. '{$NF++;print}')
  NEW_SEQUENCE=$(echo ${CURRENT_SEQUENCE} | awk -F. -v OFS=. '{$NF++;print}')
  echo "CURRENT_VERSION: $CURRENT_VERSION CURRENT_SEQUENCE: $CURRENT_SEQUENCE NEW_VERSION: $NEW_VERSION NEW_SEQUENCE: $NEW_SEQUENCE"
  docker exec -ti cli-gsma cli/utils.sh setupChaincode $NEW_SEQUENCE $NEW_VERSION
  docker exec -ti cli-dtag cli/utils.sh setupChaincode $NEW_SEQUENCE $NEW_VERSION
  docker exec -ti cli-tmus cli/utils.sh setupChaincode $NEW_SEQUENCE $NEW_VERSION
  docker exec -ti cli-gsma cli/utils.sh commitChaincode $NEW_SEQUENCE $NEW_VERSION
}

function setup_dtag() {
  docker exec -ti --user nomad webapp-dtag node setup.js
  docker exec -ti cli-dtag cli/script.sh setup
}

function setup_tmus() {
  docker exec -ti --user nomad webapp-tmus node setup.js
  docker exec -ti cli-tmus cli/script.sh setup
}

function setup_gsma() {
  docker exec -ti cli-gsma cli/script.sh setup
}

function setup() {
  if [ $# -eq 0 ]; then
    setupChannel
    setupChaincodes
    setupWebapp
  else
    case $1 in
      dtag)
        setup_dtag
        ;;
      tmus)
        setup_tmus
        ;;
      gsma)
        setup_gsma
        ;;
      *)
        echo "Unknow organization"
        exit 1
    esac
  fi
}

function setup_explorer() {
  docker exec explorer-db /bin/bash /opt/createdb.sh
  sleep 3s
  docker exec explorer /bin/sh -c 'cd explorer && ./start.sh'
}

function package_chaincode() {
  docker exec -ti cli-$1 cli/script.sh package-chaincode $1 $2 $3 $4
}

function install_chaincode() {
  docker exec -ti cli-$1 cli/script.sh install-chaincode $1 $2 $3
}

function approve_chaincode() {
  docker exec -ti cli-$1 cli/script.sh approve-chaincode $1 $2 $3 $4
}

function commit_chaincode() {
  docker exec -ti cli-$1 cli/script.sh commit-chaincode $1 $2 $3 $4
}

function upgrade_chaincode() {
  docker exec -ti cli-dtag cli/script.sh upgrade-chaincode $3 $4
  docker exec -ti cli-tmus cli/script.sh upgrade-chaincode $3 $4
  docker exec -ti cli-gsma cli/script.sh upgrade-chaincode $3 $4
  docker exec -ti cli-dtag cli/script.sh commit-chaincode $1 $2 $3 $4
}

function update_crl() {
  docker exec -ti cli-$1 cli/script.sh update-crl $1
}

#function upgrade_webapp() {
#  docker-compose up -d --no-deps $1
#  docker exec -ti --user root $1 node setup.js
#}

function info() {
  docker exec -ti cli-$1 cli/script.sh info $1 $2
}

function invoke() {
  docker exec -ti cli-$1 cli/script.sh invoke $1 $2
}

function query() {
  docker exec -ti cli-$1 cli/script.sh query $1 $2
}

function tty() {
  CONTAINER_ID_OR_NAME=$1
  docker exec -ti --user root $CONTAINER_ID_OR_NAME /bin/sh
}


function build() {
  HASH=$(cat .git/modules/blockchain-adapter/HEAD || echo "NO_HEAD" | head -1 | cut -f1)
  docker-compose build --build-arg BSA_COMMIT_HASH="$HASH" $1 || exit 1
}

function rebuild() {
  NAME=$1
  build $NAME
  docker-compose up -d --remove-orphans
}

function up() {
  docker-compose up
}

function down() {
  filter='hyperledger\|dev-peer\|mysql\|nginx\|restadapter\|offchain\|blockchain-adapter'
  #docker stop $(docker ps -a -q)
  docker stop $(docker ps -a | grep $filter | awk '{print $1}')
  #docker kill $(docker ps -q)
  docker kill $(docker ps -a | grep $filter | awk '{print $1}')
  #docker rm $(docker ps -aq)
  docker rm $(docker ps -a | grep $filter | awk '{print $1}')
  #docker rmi $(docker images dev-* -q)
  docker rmi $(docker images dev-peer* -q)
  docker volume prune --force
}
# Print help when there are no arguments
if [ "$#" -eq 0 ]
then
  printHelp >&2
  exit 1
fi
# Check if the function exists (bash specific)
if declare -f "$1" > /dev/null
then
  # call arguments verbatim
  "$@"
else
  # Show a helpful error
  echo "'$1' is not a known function name" >&2
  printHelp >&2
  exit 1
fi
