This demo shows the extended abilities of the replacement plugin
described in [this document](https://docs.google.com/presentation/d/1bXVwY11LIS6awzifOrcLsXs_b4l7V8Yl04FqOLGybuc/edit#slide=id.g1f87997393_0_782).
The multilayer structure is taken from [another document](https://docs.google.com/presentation/d/1gCAIsETGFYjVim0ChEQHmDwTtrJWbcFiBbFGDN7gQXA/edit#slide=id.g1f87997393_0_782), was a bit modified to comply with multilayer approach selected in Airship2.0 and used just as an example to demonstrate how to use the new abilities in that scenario.

The main yaml document that we're working here with is [this secret](manifests/function/ephemeral/secret.yaml) that can be used by BaremetalHost.
In this example we'll show how to modify it's parameters:
 * versions of [docker-ce, docker-ce-cli](manifests/function/ephemeral/secret.yaml#L37), [kubelet, kubeadm, kubectl](manifests/function/ephemeral/secret.yaml#L40)
 * [ssh users credentials](manifests/function/ephemeral/secret.yaml#L14)
 * [set of k8s CA and keys](manifests/function/ephemeral/secret.yaml#L44) in different sections. Please pay attention, that certificate-authority-data is also used in /etc/kubernetes/pki/ca.crt - and it will allow us to show deduplication approach based on that.
 
Let's define the parameters we want to modify in the function [catalogue.yaml](manifests/function/ephemeral/catalogue.yaml). Here is the part of this document:

```
...
versions:
  docker-ce: 19.03.12_
  docker-ce-cli: 19.03.12_
  kubelet: 1.18.6-00_
  kubeadm: 1.18.6-00_
  kubectl: 1.18.6-00_
creds:
  users: |
    root:deploY!K8s_
    deployer:deploY!K8s_
...
```

Please note that the values were intentionally appened with `_`, so it would be possible to see that replacement really works. In real scenario this document should contain the default values for all parameters.

These both resources present in [kustomization.yaml](manifests/function/ephemeral/kustomization.yaml).

The next file we want to check is the configuration of replacement transformer that will set values from the catalogue in the original secret. It has the same name as the file it modifies, but is located in the replacement catalog [here](manifests/function/ephemeral/replacements/secret.yaml).

There are several replacement objects. Let's look throuh them in order to understand how they work.

Here is the first example. Please refer to the comments to understand the Multiref and Yaml-Iin-Yaml in details:

```
- source:
    multiref:
      refs:
      - objref:
          kind: Secret
          name: node1-bmc-secret
        # Using Yaml-in-Yaml feature to get the current value of 8's element of runcmd array. 
        # It's value will be in Values[0]
        fieldref: stringData.userData|runcmd[8]
      - objref:
          name: ephemeral-catalogue
        # This is a value that we want to set instead of the current docker-ce version.
        # This value will be in Values[1]
        fieldref: versions.docker-ce 
      # Once all values are collected this template will be executed.
      # We're locating the place in the original string, generating the new string and perform replacement
      template: | 
        {{ regexReplaceAll "docker-ce . grep 19.03.12" (index .Values 0) (printf "docker-ce | grep %s" (index .Values 1)) }}
  target:
    objref:
      kind: Secret
      name: node1-bmc-secret
    # Once source value is calculated (with go-template in this case) this value is set to all fieldrefs in target
    fieldrefs: 
    - stringData.userData|runcmd[8] # please pay attention that we're putting the modified value to the same place
```

The demonstrated apporach shows how to update part of the original string with new value.

The second example will show a bit different technique:

```
- source:
    multiref:
      refs:
      - objref:
          kind: Secret
          name: node1-bmc-secret
        fieldref: stringData.userData|runcmd[=apt install -y kubelet=1.18.6-00 kubeadm=1.18.6-00 kubectl=1.18.6-00]
      - objref:
          name: ephemeral-catalogue
        fieldref: versions.kubelet
      - objref:
          name: ephemeral-catalogue
        fieldref: versions.kubeadm
      - objref:
          name: ephemeral-catalogue
        fieldref: versions.kubectl
      template: |
        {{ regexReplaceAll "kubectl=1.18.6-00" (regexReplaceAll "kubeadm=1.18.6-00" (regexReplaceAll "kubelet=1.18.6-00" (index .Values 0) (printf "kubelet=%s" (index .Values 1))) (printf "kubelet=%s" (index .Values 2))) (printf "kubelet=%s" (index .Values 3)) }}
  target:
    objref:
      kind: Secret
      name: node1-bmc-secret
    fieldrefs:
    - stringData.userData|runcmd[=apt install -y kubelet=1.18.6-00 kubeadm=1.18.6-00 kubectl=1.18.6-00]
```

In the second example I get access to the element I want to modify using `[=<value>]` construction. In this case that means that
we must specify the exact string we want to modify. We still perform regexReplaceAll to find and change all needed values, but we may consider just
building the new string with pattern for some similar cases. Access by index used in the first example also has drawback -
due to the changes in the original yaml it may be that the index may be incorrect.

Please use this to run the demo(make sure that docker is installed before):

```
$ ./run.sh
```
