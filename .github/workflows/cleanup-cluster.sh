#!/usr/bin/env bash

set -o nounset

source ${FATS_DIR}/.configure.sh

pkill kubectl
riff streaming kafka-gateway delete franz --namespace $NAMESPACE
kubectl delete ns $NAMESPACE

echo "cleanup riff streaming runtime"
kapp delete -y -n apps -a riff-streaming-runtime
kapp delete -y -n apps -a keda

if [ $GATEWAY = kafka ] ; then
    echo "cleanup kafka"
    kapp delete -y -n apps -a internal-only-kafka
fi
if [ $GATEWAY = pulsar ] ; then
    echo "cleanup pulsar"
    kapp delete -y -n apps -a internal-only-pulsar
fi

echo "cleanup riff build"
kapp delete -y -n apps -a riff-build
kapp delete -y -n apps -a riff-builders
kapp delete -y -n apps -a kpack

echo "cleanup cert manager"
kapp delete -y -n apps -a cert-manager

source ${FATS_DIR}/cleanup.sh
