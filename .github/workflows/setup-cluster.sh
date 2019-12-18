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
fats_retry kapp deploy -y -n apps -a cert-manager -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/cert-manager.yaml -y

echo "install riff build"
kapp deploy -y -n apps -a kpack -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/kpack.yaml -y
kapp deploy -y -n apps -a riff-builders -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/riff-builders.yaml -y
kapp deploy -y -n apps -a riff-build -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/riff-build.yaml -y

echo "install kafka"
kapp deploy -y -n apps -a kafka -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/kafka.yaml -y

echo "installing riff streaming runtime"
kapp deploy -n apps -a keda -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/keda.yaml -y
kapp deploy -n apps -a riff-streaming-runtime -f https://storage.googleapis.com/projectriff/charts/uncharted/${riff_version}/riff-streaming-runtime.yaml -y

riff streaming kafka-provider create franz --bootstrap-servers kafka.kafka:9092 --namespace $NAMESPACE
kubectl wait --for=condition=Ready "pod/$(kubectl -n $NAMESPACE get pod -lstreaming.projectriff.io/kafka-provider-gateway -otemplate --template="{{(index .items 0).metadata.name}}")" -n $NAMESPACE --timeout=60s
kubectl -n $NAMESPACE port-forward "svc/$(kubectl -n $NAMESPACE get svc -lstreaming.projectriff.io/kafka-provider-gateway -otemplate --template="{{(index .items 0).metadata.name}}")" "6565:6565" &
