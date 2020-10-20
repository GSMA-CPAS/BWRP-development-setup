# Local Network

## Requirements

* Docker & Docker-Compose
* Linux or MAC OS X
* OpenSSL 1.1.1

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

### (3) Update ``/etc/hosts``. Replace 192.168.2.119 with your host ip

<pre>
192.168.2.119  dtag.poc.com.local
192.168.2.119  tmus.poc.com.local
</pre>

### (4) Build required images

<pre>
$ ./nomad.sh build
</pre>

Possible issues:
-  ERROR: Pool overlaps with other one on this address space / or other resources on docker/
Free the needed resource or change the used one, examle prune/free the used network or change the netowrk range in docker-compose.yaml.
-  ERROR: Service 'blockchain-adapter-tmus' failed to build: Get https://registry-1.docker.io/v2/: dial tcp: lookup registry-1.docker.io on [::1]:53: read udp [::1]:41959->[::1]:53: read: connection refused
edit /etc/resolve.comf to include:
nameserver 8.8.8.8

### (5) Launch network

<pre>
$ ./nomad.sh up
</pre>

Wait until cluster stable appears

### (6) Setup channel and chaincode

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
BSA_DTAG="localhost:8081"
BSA_TMUS="localhost:8082"
</pre>

## upgrade Chaincode

<pre>
$ cd chaincode
$ git pull
$ ./nomad.sh upgradeChaincodes
</pre>
