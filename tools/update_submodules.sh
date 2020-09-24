#!/bin/bash
set -e
if [ $# -ne 1 ]; then
    echo "> usage: $0 <BRANCH>"
    exit 1
fi

BRANCH=$1

BASE="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null 2>&1 && pwd )"

DIRS=( "blockchain-adapter" "chaincode" "offchain_db_adapter" )
COUNT=0

for dir in "${DIRS[@]}"
do
  echo "> processing $dir, checking out $BRANCH"
  cd $BASE/../$dir
  git pull
  # use branch if existing:
  echo -ne "[$dir] "
  RES=$(git ls-remote --heads origin $BRANCH | wc -l)
  if [ $RES -gt 0 ]; then
    echo "using branch $BRANCH" 
    git checkout $BRANCH
    COUNT=$((COUNT+1))
  else 
    echo "branch $BRANCH not found, will use master"
    git checkout master
  fi
  git pull
done

if [ $COUNT == 0 ]; then
	echo "> checkout failed, branch $BRANCH was not found on any repo"
	exit 1
fi

echo "> done, using $BRANH in $COUNT repos"

