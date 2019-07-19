#!/bin/bash

elastic=~/Desktop/elasticsearch/elasticsearch-6.5.0/
kibana=~/Desktop/elasticsearch/kibana-6.5.0-darwin-x86_64/

if [ $# -eq 1 ]; then
    elastic=$1
elif [ $# -eq 2 ]; then
    kibana=$2
elif [ $# -gt 2 ]; then
    echo "Usage: start_service.sh <elasticsearch_path> <kibana_path>"
    exit 1
fi

sh ${elastic}bin/elasticsearch &
sleep 30
sh ${kibana}bin/kibana &
sleep 15