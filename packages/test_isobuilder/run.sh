#!/bin/sh
if [ ! -f ./kpt ]; then
  curl -fsSL https://github.com/GoogleContainerTools/kpt/releases/download/v0.34.0/kpt_linux_amd64-0.34.0.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

#KUSTOMIZE_PLUGIN_HOME=$(pwd)/manifests ./kustomize build --enable_alpha_plugins manifests/site/ephemeral/bootstrap/  > output.yaml
#diff manifests/function/ephemeral/secret.yaml output.yaml > changes.diff

[ -d manifests/kustomize/plugin/ ] || mkdir -p manifests/kustomize/plugin/
[ -d workdir/ ] || mkdir workdir/

# build image
docker build -f build/kustomize_build_dind.Dockerfile ./ -t quay.io/aodinokov/kustomize_build_dind:0.0.1

echo 'kind: x' |
  ./kpt fn run \
    --fn-path fn/00_ephemeral_build_iso/fn.yaml \
    --network \
    --mount type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock,rw=true \
    --mount type=bind,src="$(pwd)"/manifests/,dst=/manifests/ \
    --mount type=bind,src="$(pwd)"/workdir/,dst=/workdir/,rw=true
