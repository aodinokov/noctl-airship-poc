# Templater function

This is an example of implementing a templater function.
The feature set and resource model is similar Airship2 [templater](https://opendev.org/airship/airshipctl/src/branch/master/pkg/document/plugin/templater).

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize config run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kustomize config run local-resource/

The generated resources will appear in local-resource/

```
$ cat local-resource/*

apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-1
spec:
  bootMACAddress: 00:aa:bb:cc:dd

apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-2
spec:
  bootMACAddress: 00:aa:bb:cc:ee
...
```
