#!/usr/bin/env bash

set -o nounset

source ${FATS_DIR}/.configure.sh

pkill kubectl
riff streaming kafka-gateway delete franz --namespace $NAMESPACE
kubectl delete ns $NAMESPACE

echo "cleanup riff streaming runtime"
kapp delete -y -n apps -a riff-streaming-runtime
kapp delete -y -n apps -a keda

echo "cleanup kafka"
kapp delete -y -n apps -a kafka

echo "cleanup riff build"
kapp delete -y -n apps -a riff-build
kapp delete -y -n apps -a riff-builders
kapp delete -y -n apps -a kpack

echo "cleanup cert manager"
kapp delete -y -n apps -a cert-manager

source ${FATS_DIR}/cleanup.sh
