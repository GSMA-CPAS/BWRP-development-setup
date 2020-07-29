# Local Network

## Requirements

* Docker & Docker-Compose
* Linux or MAC OS X

## Installation

### (1) Clone webbapp repo

<pre>
$ cd nomad
$ git clone ssh://git@git.trilobyte-se.de/nomad/nomad-gsma-atomic/webapp.git
</pre>

### (2) Build webapp docker image

<pre>
$ cd nomad/webapp
$ docker build --no-cache -t webapp:1.0.0 .
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
WEBAPP_VERSION=1.0.0

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

### (5) Update ``/etc/hosts``. Replace 192.168.2.119 with your host ip

<pre>
192.168.2.119  dtag.poc.com.local
192.168.2.119  tmus.poc.com.local
</pre>

### (6) Launch network

<pre>
$ cd network-local
$ docker-compose up
</pre>

### (7) Setup network and webapp

<pre>
$ cd network-local
$ ./nomad.sh setup
</pre>

### (8) Open webapp

**DTAG**

Url: https://dtag.poc.com.local

**TMUS**

Url: https://tmus.poc.com.local
