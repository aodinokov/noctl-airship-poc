FROM node:lts-alpine as builder

RUN apk add bash curl git && apk update

ARG CLUSTERCTL_VERSION="v0.3.7"
RUN curl -fsSL -o /usr/local/bin/clusterctl https://github.com/kubernetes-sigs/cluster-api/releases/download/${CLUSTERCTL_VERSION}/clusterctl-linux-amd64 && \
    chmod +x /usr/local/bin/clusterctl

#RUN curl -fsSL -o /usr/local/bin/kpt https://storage.googleapis.com/kpt-dev/latest/linux_amd64/kpt && \
#    chmod +x /usr/local/bin/kpt

RUN mkdir -p /home/node/app && \
    chown -R node:node /home/node/app

USER node

WORKDIR /home/node/app

# Install dependencies and cache them.
COPY --chown=node:node package*.json ./
# Make rw package work
COPY --chown=node:node @types @types

RUN npm ci --ignore-scripts

# Build the source.
COPY --chown=node:node tsconfig.json .
COPY --chown=node:node src src
RUN npm run build && \
    npm prune --production && \
    rm -r src tsconfig.json

#############################################

FROM node:lts-alpine

RUN apk add git

# Run as non-root user as a best-practices:
# https://github.com/nodejs/docker-node/blob/master/docs/BestPractices.md
USER node

WORKDIR /home/node/app

COPY --from=builder /home/node/app /home/node/app

COPY --from=builder /usr/local/bin /usr/local/bin

ENV PATH /usr/local/bin:$PATH

ENTRYPOINT ["node", "/home/node/app/dist/clusterctl_run.js"]
