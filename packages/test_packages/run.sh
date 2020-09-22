#!/bin/sh
if [ ! -f ./kpt ]; then
  curl -fsSL https://github.com/GoogleContainerTools/kpt/releases/download/v0.34.0/kpt_linux_amd64-0.34.0.tar.gz -o x.tar.gz && tar -zxvf x.tar.gz && rm x.tar.gz
fi

# TODO: remove this
rm -rf exm01a/

[ -d exm01a/ ] || 
	./kpt pkg get https://github.com/aodinokov/noctl-airship-poc/packages/clusters/exm01a exm01a &&
	./kpt pkg sync exm01a/

cd exm01a/
[ -d cache ] || mkdir cache

## build image
#docker build -f build/kustomize_build_dind.Dockerfile ./ -t quay.io/aodinokov/kustomize_build_dind:0.0.1

cat Kptfile |
  ../kpt fn run \
    --fn-path phases/gf \
    --network \
    --mount type=bind,src=/var/run/docker.sock,dst=/var/run/docker.sock,rw=true \
    --mount type=bind,src="$(pwd)",dst=/cluster_root/ \
    --mount type=bind,src="$(pwd)"/cache/,dst=/cache/,rw=true
