#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.8.1/kustomize_v3.8.1_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

#KUSTOMIZE_PLUGIN_HOME=$(pwd)/manifests-example ./kustomize build --enable_alpha_plugins manifests-example/site/test-workload/target/workload/ > output.yaml

# example 1: calling as a function 
#./sops keyservice --net tcp --addr 0.0.0.0:8765
# or mount
# wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc && gpg --import sops_functional_tests_key.asc
rm -rf lr lgnupg
rm sops_functional_tests_key.asc

wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
mkdir lgnupg && GNUPGHOME=$PWD/lgnupg gpg --import sops_functional_tests_key.asc && sudo chown -R nobody lgnupg
cp -r local-resource lr && HOME='' ./kustomize fn run lr/ --mount type=bind,source=$PWD/lgnupg/,target=/.gnupg/,rw=true
