# Local Network

## Requirements

* Docker & Docker-Compose
* Linux or MAC OS X

## Installation

### (1) Clone restadapter repo

<pre>
$ cd nomad
$ git clone ssh://git@git.trilobyte-se.de/nomad/nomad-gsma-atomic/restadapter.git
</pre>

### (2) Build restadapter docker image

<pre>
$ cd nomad/restadapter
$ docker build --no-cache -t restadapter:1.0.0 .
</pre>

### (3) Clone local-network repo

<pre>
$ cd nomad
$ git clone ssh://git@git.trilobyte-se.de/nomad/nomad-gsma-atomic/network-local.git
</pre>

### (4) Create ``.env`` file in network-local (example .env-template)

<pre>
$ cd nomad/network-local
$ vi .env
</pre>

<pre>
COMPOSE_PROJECT_NAME=nomad
HLF_VERSION=2.1.0
HLF_CA_VERSION=1.4.6
COUCHDB_VERSION=0.4.18
REST_ADAPTER_VERSION=1.0.0

CLIENTAUTHREQUIRED=true

DTAG_COUCHDB_USER=nomad
DTAG_COUCHDB_PASSWORD=Grd5EfTg!dd
DTAG_CA_ADMIN_ENROLLMENT_SECRET=73rwbu37rb37ruwbrw3r
DTAG_CA_USER_ENROLLMENT_SECRET=fdjfh74bwbs74rwjrb
DTAG_MYSQL_ROOT_PASSWORD=Ac3d!dewD
DTAG_MYSQL_USER=nomad
DTAG_MYSQL_PASSWORD=Fe3gtZ6!s4Fe

TMUS_COUCHDB_USER=nomad
TMUS_COUCHDB_PASSWORD=Grd5EfTg!dd
TMUS_CA_ADMIN_ENROLLMENT_SECRET=73rwbu37rb37ruwbrw3r
TMUS_CA_USER_ENROLLMENT_SECRET=fdjfh74bwbs74rwjrb
TMUS_MYSQL_ROOT_PASSWORD=Ac3d!dewD
TMUS_MYSQL_USER=nomad
TMUS_MYSQL_PASSWORD=Fe3gtZ6!s4Fe

GSMA_COUCHDB_USER=nomad
GSMA_COUCHDB_PASSWORD=Grd5EfTg!dd
</pre>

### (5) Launch network

<pre>
$ cd network-local
$ docker-compose up
</pre>

### (6) Setup network and restadapter

Open new tab in the current terminal

<pre>
$ cd network-local
$ ./nomad.sh setup
</pre>

## Test offchain communication

Query SetSQLDBConn to configure the chaincode to read/write to mysql database.

<pre>
$ ./nomad.sh query dtag peer0
$ ./nomad.sh query tmus peer0
</pre>

Enter CLI container of organization DTAG:

<pre>
$ ./nomad.sh tty cli-dtag
</pre>

Install curl:

<pre>
$ apk --no-cache add curl
</pre>

SetData:

<pre>
$ curl -v -X POST "http://restadapter-dtag:3000/api/v1/offchain/setData/abcd?org=TMUS" -d'{"hello":"world"}'
</pre>

GetData:

<pre>
$ curl -v -X GET "http://restadapter-dtag:3000/api/v1/offchain/getData/abcd?org=TMUS&val=true"
</pre>

VerifyRemote:

<pre>
$ curl -v -X GET "http://restadapter-dtag:3000/api/v1/offchain/verifyRemote/abcd?org=TMUS"
</pre>

## Create new chaincode package (tar.gz)

Example: create new chaincode package (v1.1.0) for organization DTAG. It will be stored in ``/organizations/dtag/cli/``. This package can later be used for all other organization.

<pre>
$ ./nomad.sh tty cli-dtag
$ cd /opt/gopath/src/github.com/chaincode/offchain/1.0.0
$ GO111MODULE=on go mod vendor
$ cd /opt/gopath/src/github.com/hyperledger/fabric/peer
$ peer lifecycle chaincode package cli/offchain-v1.0.0.tar.gz --path /opt/gopath/src/github.com/chaincode/offchain/1.0.0/ --label offchain_v1.1.0
</pre>
