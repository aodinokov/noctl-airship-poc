#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.8.1/kustomize_v3.8.1_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

rm sops_functional_tests_key.asc
wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc

echo build 1
KUSTOMIZE_PLUGIN_HOME=$(pwd) SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) ./kustomize build --enable_alpha_plugins site/site1 > output1.yaml

echo regenerating secrets
# this should be substituted with https://review.opendev.org/c/airship/airshipctl/+/765593 with kustomizeSinkOutputDir
KUSTOMIZE_PLUGIN_HOME=$(pwd) SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) SOPS_PGP_FP='FBC7B9E2A4F9289AC0C1D4843D16CEE4A27381B4' ./kustomize build --enable_alpha_plugins type/type1/secrets_regenerator/ | ./kustomize fn sink site/site1/secrets/generated/

echo build 2
KUSTOMIZE_PLUGIN_HOME=$(pwd) SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) ./kustomize build --enable_alpha_plugins site/site1 > output2.yaml

diff output1.yaml output2.yaml
