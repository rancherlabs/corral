#!/bin/bash
set -ex

CORRAL_listvar="${CORRAL_listvar}"
CORRAL_mapvar="${CORRAL_mapvar}"

echo listvar ${CORRAL_listvar}
echo mapvar ${CORRAL_mapvar}

echo "corral_set listvar2=${CORRAL_listvar}"
echo "corral_set mapvar2=${CORRAL_mapvar}"

echo "corral_set numbervar=2"
echo "corral_set stringvar=abc"