#!/bin/bash
set -e

DIRS=( "blockchain-adapter" "chaincode" "offchain_db_adapter" )
for dir in "${DIRS[@]}"
do
  echo "> processing $dir"
  cd $dir
  git checkout master
  git pull
  cd ..
done
