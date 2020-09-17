#!/bin/bash
set -e

if [ $# -ne 1 ]; then
    echo "> usage: $0 branch"
    exit 1
fi

BASE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

BRANCH=$1

cd $BASE/..

git checkout $BRANCH

$BASE/update_submodules.sh $BRANCH

echo "done. now run"
echo ""
echo "sudo ./nomad.sh down"
echo "sudo docker-compose up"
echo "<wait until all pods are up and running>"
echo "sudo ./nomad.sh setup"
echo ""
echo ""
