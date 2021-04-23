#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.1.2/kustomize_v4.1.2_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

rm keys.txt
wget https://raw.githubusercontent.com/mozilla/sops/master/age/keys.txt

SOPS_IMPORT_AGE=$(cat keys.txt) SOPS_AGE_RECIPIENTS='age1yt3tfqlfrwdwx0z0ynwplcr6qxcxfaqycuprpmy89nr83ltx74tqdpszlw' ./kustomize build --enable-alpha-plugins manifests/ #> output.yaml

exit 0

[ -d test ] && rm -rf test
mkdir test
cat output.yaml | ./kustomize fn sink test

