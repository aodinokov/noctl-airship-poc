#!/bin/sh
if [ ! -f ./kustomize ]; then
  curl -fsSL https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize%2Fv3.8.1/kustomize_v3.8.1_linux_amd64.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz

  ./kustomize build helloWorld > 1.yaml
   echo "- arbitraryCleanup.yaml" >> helloWorld/kustomization.yaml
  ./kustomize build helloWorld > 2.yaml
fi

diff 1.yaml 2.yaml
