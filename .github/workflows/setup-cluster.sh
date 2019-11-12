#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

source ${FATS_DIR}/start.sh

source ${FATS_DIR}/macros/helm-init.sh

helm repo add projectriff https://projectriff.storage.googleapis.com/charts/releases
helm repo update
helm install projectriff/cert-manager --name cert-manager --devel --wait
sleep 5
wait_pod_selector_ready app=cert-manager cert-manager
wait_pod_selector_ready app=webhook cert-manager

source ${FATS_DIR}/macros/no-resource-requests.sh

helm install projectriff/riff --name riff --devel --wait \
  --set cert-manager.enabled=false \
  --set tags.core-runtime=false \
  --set tags.knative-runtime=false \
  --set tags.streaming-runtime=true

helm repo add incubator http://storage.googleapis.com/kubernetes-charts-incubator
helm repo update
helm install --name my-kafka incubator/kafka --set replicas=1,zookeeper.replicaCount=1,zookeeper.env.ZK_HEAP_SIZE=128m --namespace $NAMESPACE --wait

riff streaming kafka-provider create franz --bootstrap-servers my-kafka:9092 --namespace $NAMESPACE
kubectl wait --for=condition=Ready "pod/$(kubectl -n $NAMESPACE get pod -lstreaming.projectriff.io/kafka-provider-gateway -otemplate --template="{{(index .items 0).metadata.name}}")" -n $NAMESPACE
kubectl -n $NAMESPACE port-forward "svc/$(kubectl -n $NAMESPACE get svc -lstreaming.projectriff.io/kafka-provider-gateway -otemplate --template="{{(index .items 0).metadata.name}}")" "6565:6565" &
