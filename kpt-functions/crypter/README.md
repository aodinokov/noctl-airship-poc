# Crypter function

This is an example of implementing a crypter function.
It allows to decrypt/encrypt/reencrypt(rotate) secrets in yamls. Function can be used as kpt-function or kustomize plugin.

By default it decypts the data, listed in refs, but it has a field `operation` that can in the following states: `decrypt`(default), `encrypt` or `rotate`.
First 2 values require password to be set: either as a value of another field `password`, or using env variable `crypter_password`. Env variable can be used to override
the set in the yaml value.

In case of `rotate` the function required an additional parameter `oldPassword` or `crypter_old_password` env variable.

If the password(s) isn't/aren't set the function will panic and fail.

The function also has dry-run mode that can be used to emulate behavior without knowing the actual password. In case of dry-run the fields will get generated values.

The function can be also called as a docker-container to encrypt/decrypt fields as a command-line tool, see:

	docker run quay.io/aodinokov/crypter-default:v0.0.1 config-function --help

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    crypter_password=testpass kustomize config run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kustomize config run local-resource/

The resources in local-resource/ will be decrypted

```
$ cat local-resource/*

apiVersion: airshipit.org/v1alpha1
kind: Crypter
metadata:
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/crypter-default:v0.0.1
refs:
- objref:
    kind: VariableCatalogue
    name: host-catalogue
  fieldrefs:
  - hosts.m3.node01.bmcPassword
  - hosts.m3.node02.bmcPassword
---
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: host-catalogue
hosts:
  m3:
    node01:
      macAddress: 52:54:00:b6:ed:31
      bmcAddress: redfish+http://10.23.25.1:8000/redfish/v1/Systems/air-target-1
      bmcUsername: root
      bmcPassword: r00tme
      ipAddresses:
        oam-ipv4: 10.23.25.102
        pxe-ipv4: 10.23.24.102
      macAddresses:
        oam: 52:54:00:9b:27:4c
        pxe: 52:54:00:b6:ed:31
    node02:
      macAddress: 00:3b:8b:0c:ec:8b
      bmcAddress: redfish+http://10.23.25.2:8000/redfish/v1/Systems/air-target-2
      bmcUsername: username
      bmcPassword: password
      ipAddresses:
        oam-ipv4: 10.23.25.101
        pxe-ipv4: 10.23.24.101
```
