This demo shows how [fn plugins](https://github.com/kubernetes-sigs/kustomize/tree/master/api/internal/plugins/fnplugin) work in kustomize 3.8.1
It's based on this [PR](https://review.opendev.org/#/c/735033/25) and the only changes that were made were: 


for replacement transformer added:

```
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/replacement-default:v0.0.1
```

in all documents that had kind ReplacementTransformer:

```
$ grep -Rni 'ReplacementTransformer'
manifests/function/hostgenerator-m3/replacements/hosts.yaml:4:kind: ReplacementTransformer
```

for templater generator added:

```
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/templater-default:v0.0.1
```

in all documents that had kind Templater:

```
$ grep -Rni 'Templater'
manifests/function/hostgenerator-m3/hosttemplate.yaml:2:kind: Templater
``` 

Please read [this](../test_fn_replacement/README.md) to get more information how fnplugins work.

This demo works in the following way:
[manifests/site/test-site/ephemeral/controlplane/hostgenerator/kustomization.yaml](manifests/site/test-site/ephemeral/controlplane/hostgenerator/kustomization.yaml) collects 
the template from [manifests/function/hostgenerator-m3](manifests/function/hostgenerator-m3)
the information about all hosts from [manifests/site/test-site/shared/catalogues/](manifests/site/test-site/shared/catalogues/)
the information about what hosts to generate from local file [host-generation.yaml](manifests/site/test-site/ephemeral/controlplane/hostgenerator/host-generation.yaml)
uses replacement to concat that info to the Templater with required data:

```
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/templater-default:v0.0.1
    config.kubernetes.io/path: templater_m3-host-template.yaml
  name: m3-host-template
template: |-
  {{- $envAll := . }}
  {{- range .hostsToGenerate }}
  {{- $hostName := . }}
  {{- $host := index $envAll.hosts $hostName }}
  ---
  apiVersion: metal3.io/v1alpha1
  kind: BareMetalHost
  metadata:
    annotations:
    labels:
    name: {{ $hostName }}
  spec:
    online: false
    bootMACAddress: {{ $host.macAddress }}
    networkData:
      name: {{ $hostName }}-network-data
      namespace: default
    bmc:
      address: {{ $host.bmcAddress }}
      credentialsName: {{ $hostName }}-bmc-secret
  ---
  apiVersion: v1
  kind: Secret
  metadata:
    name: {{ $hostName }}-bmc-secret
  data:
    username: {{ $host.bmcUsername | b64enc }}
    password: {{ $host.bmcPassword | b64enc }}
  type: Opaque
  ---
  apiVersion: v1
  kind: Secret
  metadata:
    name: {{ $hostName }}-network-data
  stringData:
    networkData: |
      links:
        {{- range $envAll.commonNetworking.links }}
      -
  {{ toYaml . | indent 6 }}
        {{- if $host.macAddresses }}
        ethernet_mac_address: {{ index $host.macAddresses .id }}
        {{- end }}
        {{- end }}
      networks:
        {{- range $envAll.commonNetworking.networks }}
      -
  {{ toYaml . | indent 6 }}
        ip_address: {{ index $host.ipAddresses .id }}
        {{- end }}
      services:
  {{ toYaml $envAll.commonNetworking.services | indent 6 }}
  type: Opaque

  {{ end -}}
values:
  commonNetworking:
    links:
    - id: oam
      mtu: "1500"
      name: enp0s3
      type: phy
    - id: pxe
      mtu: "1500"
      name: enp0s4
      type: phy
    networks:
    - id: oam-ipv4
      link: oam
      netmask: 255.255.255.0
      routes:
      - gateway: 10.23.25.1
        netmask: 0.0.0.0
        network: 0.0.0.0
      type: ipv4
    - id: pxe-ipv4
      link: pxe
      netmask: 255.255.255.0
      type: ipv4
    services:
    - address: 8.8.8.8
      type: dns
    - address: 8.8.4.4
      type: dns
  hosts:
    node01:
      bmcAddress: redfish+http://10.23.25.1:8000/redfish/v1/Systems/air-target-1
      bmcPassword: r00tme
      bmcUsername: root
      ipAddresses:
        oam-ipv4: 10.23.25.102
        pxe-ipv4: 10.23.24.102
      macAddress: 52:54:00:b6:ed:31
      macAddresses:
        oam: 52:54:00:9b:27:4c
        pxe: 52:54:00:b6:ed:31
    node02:
      bmcAddress: redfish+http://10.23.25.2:8000/redfish/v1/Systems/air-target-2
      bmcPassword: password
      bmcUsername: username
      ipAddresses:
        oam-ipv4: 10.23.25.101
        pxe-ipv4: 10.23.24.101
      macAddress: 00:3b:8b:0c:ec:8b
  hostsToGenerate:
  - node01
```
[cleanup.yaml](manifests/site/test-site/ephemeral/controlplane/hostgenerator/cleanup.yaml) removes all catalogues and keeps only this resource.

[manifests/site/test-site/ephemeral/controlplane/nodes/kustomization.yaml](manifests/site/test-site/ephemeral/controlplane/nodes/kustomization.yaml) calls the previous kustomization from generator section:

```
generators:
  - ../hostgenerator
```

and that makes kustomize to call Templater that generates the resources using template and values.
In addition this file adds label `airshipit.org/k8s-role: controlplane-host`

All other resources were removed from the main file [manifests/site/test-site/ephemeral/controlplane/kustomization.yaml](manifests/site/test-site/ephemeral/controlplane/kustomization.yaml) that is used to do kustomize build.

To run the example just install docker and call ./run.sh. The script will download the needed version of kustomize and put the resulting resources to output.yaml.
During the first run it will be possible to see how docker pulls the needed function images.
You can find the resulting output in the [expected_output.yaml](expected_output.yaml).
