#!/bin/bash
set -e -o errexit

if ! command -v jq &> /dev/null
then
    echo "please install \"jq\" to continue."
    exit
fi
. ./.env

function get_request {
    RET=$(curl -s -S -X GET -H "Content-Type: application/json" $1)
    echo ${RET}
    echo $RET | grep -i "error" > /dev/null && echo $RET > /dev/stderr && exit 1 || :
}
  
function post_request {
    RET=$(curl -s -S -X POST -H "Content-Type: application/json" $1 -d"$2")
    echo ${RET}
    echo $RET | grep -i "error" > /dev/null && echo $RET > /dev/stderr && exit 1 || :
}

function put_request {
    RET=$(curl -s -S -X PUT -H "Content-Type: application/json" $1 -d"$2")
    echo ${RET}
    echo $RET | grep -i "error" > /dev/null && echo $RET > /dev/stderr && exit 1 || :
}

function put2_request {
    RET=$(curl -s -S -X PUT $1)
    echo ${RET}
    echo $RET | grep -i "error" > /dev/null && echo $RET > /dev/stderr && exit 1 || :
}


echo " 1) #################### DTAG ####################"
echo " > GET /api/v1/contracts/ [get list of All Contract on DTAG]"
RET=$(get_request "http://dtag.poc.com.local:$DTAG_COMMON_ADAPTER_PORT/api/v1/contracts/")
echo $RET | jq
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 2) #################### DTAG ####################"
echo " > POST /api/v1/contracts/ [create new DRFAT Contract]"
read -p "Please enter a name for contract : "
echo " #set header->name to \"$REPLY\""
RET=$(post_request "http://dtag.poc.com.local:$DTAG_COMMON_ADAPTER_PORT/api/v1/contracts/" "{\"header\":{\"fromMsp\":{\"mspId\":\"DTAG\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"name\":\"$REPLY\",\"toMsp\":{\"mspId\":\"TMUS\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"type\":\"contract\",\"version\":\"1.0\"},\"body\":{\"key\":\"{}\"}}")
echo $RET | jq
DTAG_DOC_ID=$(echo $RET | jq -r ."contractId")
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 3) #################### DTAG ####################"
echo " > PUT /api/v1/contracts/$DTAG_DOC_ID [edit/update DRFAT Contract]"
read -p "Please enter a name for contract : "
echo " #set header->name to \"$REPLY\""
RET=$(put_request "http://dtag.poc.com.local:$DTAG_COMMON_ADAPTER_PORT/api/v1/contracts/$DTAG_DOC_ID" "{\"header\":{\"fromMsp\":{\"mspId\":\"DTAG\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"name\":\"$REPLY\",\"toMsp\":{\"mspId\":\"TMUS\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"type\":\"contract\",\"version\":\"1.0\"},\"body\":{\"key\":\"{}\"}}")
echo $RET | jq
DTAG_DOC_ID=$(echo $RET | jq -r ."contractId")
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 4) #################### DTAG ####################"
echo " > PUT /api/v1/contracts/$DTAG_DOC_ID/send/ [send Contract]"
RET=$(put2_request "http://dtag.poc.com.local:$DTAG_COMMON_ADAPTER_PORT/api/v1/contracts/$DTAG_DOC_ID/send/")
echo $RET | jq
DTAG_DOC_ID=$(echo $RET | jq -r ."contractId")
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 5) #################### TMUS ####################"
echo " > GET /api/v1/contracts/ [get list of All Contract on TMUS]"
RET=$(get_request "http://tmus.poc.com.local:$TMUS_COMMON_ADAPTER_PORT/api/v1/contracts/")
echo $RET | jq
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 6) #################### TMUS ####################"
echo " > POST /api/v1/contracts/ [create new DRFAT Contract]"
read -p "Please enter a name for contract : "
echo " #set header->name to \"$REPLY\""
RET=$(post_request "http://tmus.poc.com.local:$TMUS_COMMON_ADAPTER_PORT/api/v1/contracts/" "{\"header\":{\"fromMsp\":{\"mspId\":\"TMUS\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"name\":\"$REPLY\",\"toMsp\":{\"mspId\":\"DTAG\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"type\":\"contract\",\"version\":\"1.0\"},\"body\":{\"key\":\"{}\"}}")
echo $RET | jq
TMUS_DOC_ID=$(echo $RET | jq -r ."contractId")
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 7) #################### TMUS ####################"
echo " > PUT /api/v1/contracts/$TMUS_DOC_ID [edit/update DRFAT Contract]"
read -p "Please enter a name for contract : "
echo " #set header->name to \"$REPLY\""
RET=$(put_request "http://tmus.poc.com.local:$TMUS_COMMON_ADAPTER_PORT/api/v1/contracts/$TMUS_DOC_ID" "{\"header\":{\"fromMsp\":{\"mspId\":\"TMUS\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"name\":\"$REPLY\",\"toMsp\":{\"mspId\":\"DTAG\",\"signatures\":[{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"},{\"role\":\"role\",\"name\":\"name\",\"id\":\"id\"}]},\"type\":\"contract\",\"version\":\"1.0\"},\"body\":{\"key\":\"{}\"}}")
echo $RET | jq
TMUS_DOC_ID=$(echo $RET | jq -r ."contractId")
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 8) #################### TMUS ####################"
echo " > PUT /api/v1/contracts/$TMUS_DOC_ID/send/ [send Contract]"
RET=$(put2_request "http://tmus.poc.com.local:$TMUS_COMMON_ADAPTER_PORT/api/v1/contracts/$TMUS_DOC_ID/send/")
echo $RET | jq
TMUS_DOC_ID=$(echo $RET | jq -r ."contractId")
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo


echo " 9) #################### DTAG ####################"
echo " > GET /api/v1/contracts/ [get list of All Contract on DTAG]"
RET=$(get_request "http://dtag.poc.com.local:$DTAG_COMMON_ADAPTER_PORT/api/v1/contracts/")
echo $RET | jq
echo
read -n 1 -s -r -p "Press any key to continue ..."
echo
echo "--END--"
