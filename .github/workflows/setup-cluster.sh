#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source ${FATS_DIR}/start.sh

readonly riff_version=0.5.0-snapshot

source ${FATS_DIR}/.configure.sh
kubectl create ns apps
kubectl create ns $NAMESPACE

echo "install cert manager"
fats_retry kapp deploy -y -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/release/${riff_version}/cert-manager.yaml

echo "install riff build"
kapp deploy -y -n apps -a kpack -f https://storage.googleapis.com/projectriff/release/${riff_version}/kpack.yaml
kapp deploy -y -n apps -a riff-builders -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-builders.yaml
kapp deploy -y -n apps -a riff-build -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-build.yaml

if [ $GATEWAY = kafka ] ; then
    echo "install kafka"
    kapp deploy -y -n apps -a internal-only-kafka -f https://storage.googleapis.com/projectriff/release/${riff_version}/internal-only-kafka.yaml
fi
if [ $GATEWAY = pulsar ] ; then
    echo "install kafka"
    kapp deploy -y -n apps -a internal-only-pulsar -f https://storage.googleapis.com/projectriff/release/${riff_version}/internal-only-pulsar.yaml
fi

echo "installing riff streaming runtime"
kapp deploy -y -n apps -a keda -f https://storage.googleapis.com/projectriff/release/${riff_version}/keda.yaml
kapp deploy -y -n apps -a riff-streaming-runtime -f https://storage.googleapis.com/projectriff/release/${riff_version}/riff-streaming-runtime.yaml

if [ $GATEWAY = inmemory ] ; then
  riff streaming inmemory-gateway create test --namespace $NAMESPACE --tail
fi
if [ $GATEWAY = kafka ] ; then
  riff streaming kafka-gateway create test --bootstrap-servers kafka.kafka.svc.cluster.local:9092 --namespace $NAMESPACE --tail
fi
if [ $GATEWAY = pulsar ] ; then
  riff streaming pulsar-gateway create test --service-url pulsar://pulsar.pulsar.svc.cluster.local:6650 --namespace $NAMESPACE --tail
fi

kubectl -n $NAMESPACE port-forward "$(kubectl -n $NAMESPACE get svc -lstreaming.projectriff.io/gateway=test -oname)" "6565:6565" &
