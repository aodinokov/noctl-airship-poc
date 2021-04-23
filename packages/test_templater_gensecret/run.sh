#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.1.2/kustomize_v4.1.2_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

rm sops_functional_tests_key.asc
wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc

SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) SOPS_PGP_FP='FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4' ./kustomize build --enable-alpha-plugins manifests/ #> output.yaml

exit 0

[ -d test ] && rm -rf test
mkdir test
cat output.yaml | ./kustomize fn sink test

