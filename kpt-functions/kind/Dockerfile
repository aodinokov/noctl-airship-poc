ARG RELEASE_IMAGE=alpine:3.12.0
FROM ${RELEASE_IMAGE} as release
RUN apk update && apk add curl docker-cli && rm -rf /var/cache/apk/*
ARG KUBE_LATEST_VERSION="v1.21.1"
RUN curl -L https://dl.k8s.io/release/${KUBE_LATEST_VERSION}/bin/linux/amd64/kubectl -o /bin/kubectl \
    && chmod +x /bin/kubectl
ARG KIND_LATEST_VERSION="v0.11.0"
RUN curl -L https://kind.sigs.k8s.io/dl/${KIND_LATEST_VERSION}/kind-linux-amd64 -o /bin/kind \
    && chmod +x /bin/kind
