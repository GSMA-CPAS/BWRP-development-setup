# Local Network

## Requirements

* Docker & Docker-Compose
* Linux or MAC OS X

## Installation

### (1) Clone git submodules

<pre>
$ cd BWRP-development-setup
$ git submodule update --init
</pre>

### (2) Create ``.env`` file in BWRP-development-setup (example .env-template)

<pre>
$ cp .env-template .env
$ vi .env #add passwords etc.
</pre>

### (3) Build required images

<pre>
$ docker-compose build
</pre>

### (4) Launch network

<pre>
$ docker-compose up
</pre>

Wait until cluster stable appears

### (5) Setup channel and chaincode

Open new tab in the current terminal

<pre>
$ cd BWRP-development-setup
$ ./nomad.sh setup
</pre>

Wait until chaincode is committed

## Test blockchain-adapter rest-api

Install curl and jq:
<pre>
$ apt install jq curl
</pre>

Run the test script:
<pre>
$ ./blockchain-adapter/test_query.sh
</pre>

It should finish with:  Verified OK

If it fails with hostname not found you can set the hostnames in /etc/hosts or run the script inside docker or change the hosts inside the script to localhost:
<pre>
BSA_DTAG="localhost:8080"
BSA_TMUS="localhost:8081"
</pre>

## upgrade Chaincode

<pre>
$ cd chaincode
$ git pull
$ ./nomad.sh upgradeChaincodes
</pre>
