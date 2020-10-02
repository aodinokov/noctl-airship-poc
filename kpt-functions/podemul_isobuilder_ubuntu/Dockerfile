# Copyright 2018 AT&T Intellectual Property.  All other rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

ARG FROM=ubuntu:focal

FROM quay.io/aodinokov/podemul:v0.0.1 AS podemul

FROM ${FROM} AS ubuntu
COPY --from=podemul /usr/local/bin/config-function /usr/local/bin/config-function
RUN apt-get update && apt-get install -y \
    curl\
    debootstrap \
    dosfstools \
    grub-efi-amd64-bin \
    grub-pc-bin \
    mtools \
    squashfs-tools \
    xorriso \
    && rm -rf /var/lib/apt/lists/*
RUN curl -L https://github.com/mikefarah/yq/releases/download/2.4.0/yq_linux_amd64 -o /bin/yq \
    && chmod +x /bin/yq
CMD /usr/local/bin/config-function
