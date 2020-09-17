#!/bin/bash
set -e
if [ $# -ne 1 ]; then
    echo "> usage: $0 <BRANCH>"
    exit 1
fi

BRANCH=$1

BASE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

DIRS=( "blockchain-adapter" "chaincode" "offchain_db_adapter" )

for dir in "${DIRS[@]}"
do
  echo "> processing $dir, checking out $BRANCH"
  cd $BASE/../$dir
  git pull
  git checkout $BRANCH
  git pull
done

