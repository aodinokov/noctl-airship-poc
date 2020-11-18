#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.8.1/kustomize_v3.8.1_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

rm sops_functional_tests_key.asc
wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc

KUSTOMIZE_PLUGIN_HOME=$(pwd)/manifests SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) ./kustomize build --enable_alpha_plugins manifests/ > output.yaml

[ -d test ] && rm -rf test
mkdir test
cat output.yaml | ./kustomize fn sink test

