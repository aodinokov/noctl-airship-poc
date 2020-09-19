FROM gcr.io/kpt-functions/kustomize-build:stable

USER root

RUN apk add docker

USER node
