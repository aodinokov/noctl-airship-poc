This demo shows the extended abilities of the replacement plugin
described in [this document](https://docs.google.com/presentation/d/1bXVwY11LIS6awzifOrcLsXs_b4l7V8Yl04FqOLGybuc/edit#slide=id.g1f87997393_0_782).
The multilayer structure is taken from [another document](https://docs.google.com/presentation/d/1gCAIsETGFYjVim0ChEQHmDwTtrJWbcFiBbFGDN7gQXA/edit#slide=id.g1f87997393_0_782), was a bit modified to comply with multilayer approach selected in Airship2.0 and used just as an example to demonstrate how to use the new abilities in that scenario.

The main yaml document that we're working here with is [this secret](manifests/function/ephemeral/secret.yaml) that can be used by BaremetalHost.
In this example we'll show how to modify its parameters:
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

Please note that the values were intentionally appended with `_`, so it would be possible to see that replacement really works. In real scenario this document should contain the default values for all parameters.

These both resources present in [kustomization.yaml](manifests/function/ephemeral/kustomization.yaml).

The next file we want to check is the configuration of replacement transformer that will set values from the catalogue in the original secret resource. It has the same name as the file it modifies, but is located in the replacement directory: [replacements/secret.yaml](manifests/function/ephemeral/replacements/secret.yaml).

There are several replacement objects in this yaml file. Let's look through them in order to understand how they work.

Here is the first example, please refer to the comments to understand the Multiref and Yaml-in-Yaml in details:

```
- source:
    multiref:
      # multiref object contains 2 fields: refs and template.
      # refs is an array of standard objref+fieldref pairs we had before
      # template is go-template string that may build a string based on .Values - array of values taken from refs
      refs:
      - objref:
          kind: Secret
          name: node1-bmc-secret
        # Using Yaml-in-Yaml feature (pay attention to | delimeter) to get the current value of 8's element of runcmd array. 
        # Its value will be in Values[0]
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
    # please pay attention that we're putting the modified value to the same place
    - stringData.userData|runcmd[8]
```

The demonstrated apporach shows how to update part of the original string with a new value.

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

In the second example we get access to the element we want to modify using `[=<value>]` construction. In this case that means that
we must specify the exact string we want to modify. We still perform regexReplaceAll to find and change all needed values, but we may consider just
building the new string with pattern for some similar cases. Access by index used in the first example also has drawback -
due to the changes in the original yaml it may be that the index may be incorrect. 

*Note:* the reason why we may still need both methods: by index and by value is that it's not always possible to use value - some values may have `]` that is ambiguous.

The next example shows 'deduplication':

```
- source:
    objref:
      name: ephemeral-catalogue
    fieldref: creds.certificate-authority-data
  target:
    objref:
      kind: Secret
      name: node1-bmc-secret
    fieldrefs:
    - stringData.userData|write_files.[path=/etc/kubernetes/admin.conf].content | clusters.[name=kubernetes].cluster.certificate-authority-data
    - stringData.userData|write_files.[path=/etc/kubernetes/pki/ca.crt].content
```

As it's possible to see the same value `creds.certificate-authority-data` is set to 2 fields:
 * part of `/etc/kubernetes/admin.conf` (Yaml-in-Yaml is used)
 * complete `/etc/kubernetes/pki/ca.crt`

There are more replacement objects, but they should be self-explanatory, based on the explanation of the cases already provided.

Now let's look inside [replacements/catalogue.yaml](manifests/function/ephemeral/replacements/catalogue.yaml) in the replacement directory.
This file contains strategic-merget patch to delete `ephemeral-catalogue`. This was already covered [here](../test_delete) and [here](../test_m3gen).
If you look inside the [replacements/kustomization.yaml](manifests/function/ephemeral/replacements/kustomization.yaml) from replacement directory, please note 2 things:
 * transformer configurations in `kustomization.yaml` are listed in `resources:` section. The purpose is - to be able to defer all needed transformation and call it only when the upper layer already modified the `ephemeral-catalogue` resource.
 * the order of resources: all replacements for `secret.yaml` should go before the deletion of `catalogue.yaml`. When we call this `kustomization.yaml` from the `transformers:` section these transformers will be executed in this order and the last step of `ephemeral-catalogue` deletion has to be in the bottom.
 
We've walked though all files inside [ephemeral function](manifests/function/ephemeral). Let's move to the [gating type](manifests/type/gating) that will use `ephemeral function` and will allow to change some of its parameters in 'deduplicating'-fashion. You may notice that 3 parameters: `versions.kubelet`, `versions.kubeadm` and `versions.kubectl` had the same value in `ephemeral function` catalogue. Let's define a single parameter `versions.k8s` on `gating type` level to control them all. Please refer to its [catalogue.yaml](manifests/type/gating/catalogue.yaml):

```
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: type-gating-catalogue
versions:
  k8s: 1.18.9
```

Our [kustomization.yaml](manifests/type/gating/kustomization.yaml) consists of this resourse and all resources from `ephemeral function`:

```
resources:
  - catalogue.yaml
  - ../../function/ephemeral
```

That will mean 3 resources in reality: the secret and the catalogue from `ephemeral function` plus the catalogue from `gating type`.

We define the following replacement in the corresponding [replacements/ephemeral-catalogue.yaml](manifests/type/gating/replacements/ephemeral-catalogue.yaml):

```
...
- source:
    multiref:
      refs:
      - objref:
          name: type-gating-catalogue
        fieldref: versions.k8s
      # we're adding -00 postfix for the value
      template: |-
        {{ index .Values 0 }}-00
  target:
    objref:
      name: ephemeral-catalogue
    # setting the value to the multiple fields in the target
    fieldrefs:
    - versions.kubelet
    - versions.kubeadm
    - versions.kubectl
```

We have [replacements/catalogue.yaml](manifests/type/gating/replacements/catalogue.yaml) that will remove the catalogue after using it. We also mention both transformers configuration in [replacements/kustomization.yaml](manifests/type/gating/replacements/kustomization.yaml):

```
resources:
- ephemeral-catalogue.yaml
- ../../../function/ephemeral/replacements
- catalogue.yaml
```

Please pay attention to the order of the resources. Do you remember that our resources consist of 3 resources: the secret and the catalogue from `ephemeral function` plus catalogue from `gating type`? If this 'replacements/kustomization.yaml' is called at a transformer it will perform modification of the catalogue in `ephemeral function` based on the data from the catalog in `gating type`, after that will execute all transformers from `ephemeral function`, that in our case will modify the secret fields and delete the catalogue from `ephemeral function`. And the last line here deletes the catalogue from `gating type`. Basically, this approach allows us to *incapsulate* `ephemeral function` into `gating type` in the way so the upper level above `gating type` may know about/modify the catalogue only for `gating type`.

We covered all used files in the `gating type` and now may go to the [kustomization.yaml](manifests/site/ephemeral/bootstrap/kustomization.yaml) on the site level.
To simplify things some of the fields from the original version was commented out. Here are the important parts:

```
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
- ../../../type/gating

...

transformers:
- ../../../type/gating/replacements
```

We see that we perform just 2 steps:
 * include all resources from `gating type`: as it was stated before 3 resources will appear
 * perform replacement that will configure the secret resource according to the data from the catalogues.
 
Please use this to run the demo (make sure that curl and docker are installed before):

```
$ ./run.sh
```

This script will download the required kustomize version (minimal requirement is 3.8.1) and execute the described configuration.
The `output.yaml` will contain the modified secret. The script also will create `changes.diff` to demonstrate what fields have been changed in the secret.

*Note:* if you don't have a place to run this, please refer to [expected_output.yaml](expected_output.yaml) and [expected_changes.diff](expected_changes.diff) where we copied the values we've got after run.

There are 2 possible ways to modify parameters on site level:
 * using the same approach with replacement plugin
 * using strategic merge patch (SMP) for catalogues
 
If the type layer already doesn't have any duplicated parameters, there shouldn't be any need in the replacement approach, since changing one of the parameters already has to apply changes to all dependend fields. The site-layer in most cases works only with specific values that's why the approach with SMP for the type catalogue will be simpler and quicker.

TBD: create SMP that modifies k8s version and ask the user to uncomment it in the kustomize.yaml. the versions in output.yaml must change
