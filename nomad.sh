#!/bin/bash

function printHelp() {
  echo "Usage: "
  echo "  nomad.sh setup"
  echo "  nomad.sh setup-explorer"
  echo "  nomad.sh package-chaincode <org> <peer> <version> <version>"
  echo "  nomad.sh install-chaincode <org> <peer> <version>"
  echo "  nomad.sh approve-chaincode <org> <peer> <version> <sequenz>"
  echo "  nomad.sh commit-chaincode <org> <peer> <version> <sequenz>"
  echo "  nomad.sh upgrade-chaincode <org> <peer> <version> <sequenz>"
  echo "  nomad.sh update-crl <org>"
  echo "  nomad.sh info <org> <peer>"
  echo "  nomad.sh tty <container-name>"
  echo "  nomad.sh down"
  echo
}

function setup_dtag() {
  docker exec -ti --user root webapp-dtag node setup.js
  docker exec -ti cli-dtag cli/script.sh setup
}

function setup_tmus() {
  docker exec -ti --user root webapp-tmus node setup.js
  docker exec -ti cli-tmus cli/script.sh setup
}

function setup_gsma() {
  docker exec -ti cli-gsma cli/script.sh setup
}

function setup() {
  if [ $# -eq 0 ]; then
    setup_dtag
    setup_tmus
    setup_gsma
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

function down() {
  #docker stop $(docker ps -a -q)
  docker stop $(docker ps -a | grep 'hyperledger\|dev-peer\|mysql\|nginx' | awk '{print $1}')
  #docker kill $(docker ps -q)
  docker kill $(docker ps -a | grep 'hyperledger\|dev-peer\|mysql\|nginx' | awk '{print $1}')
  #docker rm $(docker ps -aq)
  docker rm $(docker ps -a | grep 'hyperledger\|dev-peer\|mysql\|nginx' | awk '{print $1}')
  #docker rmi $(docker images dev-* -q)
  docker rmi $(docker images dev-peer* -q)
  docker volume prune --force
}

MODE=$1;shift

if [ "${MODE}" == "setup" ]; then
  setup $1
elif [ "${MODE}" == "setup-explorer" ]; then
  setup_explorer
elif [ "${MODE}" == "package-chaincode" ]; then
  package_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "install-chaincode" ]; then
  install_chaincode $1 $2 $3
elif [ "${MODE}" == "approve-chaincode" ]; then
  approve_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "commit-chaincode" ]; then
  commit_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "upgrade-chaincode" ]; then
  upgrade_chaincode $1 $2 $3 $4
elif [ "${MODE}" == "update-crl" ]; then
  update_crl $1
#elif [ "${MODE}" == "upgrade-webapp" ]; then
#  upgrade_webapp $1
elif [ "${MODE}" == "info" ]; then
  info $1 $2
elif [ "${MODE}" == "invoke" ]; then
  invoke $1 $2
elif [ "${MODE}" == "query" ]; then
  query $1 $2
elif [ "${MODE}" == "tty" ]; then
  tty $1
elif [ "${MODE}" == "down" ]; then
  down
else
  printHelp
  exit 1
fi
