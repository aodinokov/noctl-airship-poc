#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv4.1.0/kustomize_v4.1.0_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz

fi

./kustomize build --enable-alpha-plugins $1
