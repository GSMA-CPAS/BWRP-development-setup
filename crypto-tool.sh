#!/bin/bash

function printHelp() {
  echo "Usage: "
  echo "  crypto-tool.sh generate-webapp-storage-tls-certs <org> <server_name>"
  echo
}

function generate_webapp_storage_tls_certs() {
  ORG=$1
  SERVER_NAME=$2

  for file in crypto-config/peerOrganizations/${ORG}.nomad.com/tlsca/*_sk; do
    [ -f "$file" ] || break
    CA_KEY=$file
  done

  openssl ecparam -out ${SERVER_NAME}.key -name prime256v1 -genkey -noout
  openssl req -new -sha256 -key ${SERVER_NAME}.key -nodes -out ${SERVER_NAME}.csr
  openssl x509 -req -days 3650 -in ${SERVER_NAME}.csr -CA crypto-config/peerOrganizations/${ORG}.nomad.com/tlsca/tlsca.${ORG}.nomad.com-cert.pem -CAkey $CA_KEY -CAcreateserial -out ${SERVER_NAME}.crt -sha256
}

MODE=$1;shift

if [ "${MODE}" == "generate-webapp-storage-tls-certs" ]; then
  generate_webapp_storage_tls_certs $1 $2
else
  printHelp
  exit 1
fi
