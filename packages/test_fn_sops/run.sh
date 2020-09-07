#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.8.1/kustomize_v3.8.1_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

rm sops_functional_tests_key.asc
wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
KUSTOMIZE_PLUGIN_HOME=$(pwd)/manifests SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) ./kustomize build --enable_alpha_plugins manifests/base > output.yaml

# example 1: calling as a function 
#./sops keyservice --net tcp --addr 0.0.0.0:8765
# or mount
# wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc && gpg --import sops_functional_tests_key.asc

# example 2 (needs ./sops keyservice --net tcp --addr 0.0.0.0:8765 in background)
#rm -rf lr lgnupg
#rm sops_functional_tests_key.asc
#wget https://raw.githubusercontent.com/mozilla/sops/master/pgp/sops_functional_tests_key.asc
#mkdir lgnupg && GNUPGHOME=$PWD/lgnupg gpg --import sops_functional_tests_key.asc && sudo chown -R nobody lgnupg
#cp -r local-resource lr && HOME='' ./kustomize fn run lr/ --mount type=bind,source=$PWD/lgnupg/,target=/.gnupg/,rw=true

# example 3:
rm -rf lr
cp local-resource/ lr -r
SOPS_IMPORT_PGP=$(cat sops_functional_tests_key.asc) kpt fn run ./lr
# if you have your own gpg already initialized you can put keys to kpt like this:
#SOPS_IMPORT_PGP=$(gpg --armor --export-secret-keys) kpt fn run ./lr
