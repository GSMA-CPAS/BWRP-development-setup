#!/bin/bash
set -e

if [ $# -ne 1 ]; then
    echo "> usage: $0 branch"
    exit 1
fi

BASE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

cd $BASE/..

./tools/switch_to_branch.sh $1

./nomad.sh down

docker-compose build

docker-compose up -d

sleep 35

docker-compose logs -t | tail

echo "> docker should be up and running now"

./nomad.sh setup

echo "> waiting for chaincode to be up"
sleep 15

./blockchain-adapter/test_query.sh

