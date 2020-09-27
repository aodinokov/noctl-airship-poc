# Grep function

This is an example of implementing a grep function.
Grep allows filtering documents based on path, value
and etc. see [here](https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/kio/filters/grep.go#L26) for more details.

The configuration is made in form of `ConfigMap`.
all data fields must be a strings that are checked
as 'or'. each string contains several grepfilters
that are 'and' related.

## Function implementation

The function is implemented as an image (see Dockerfile).

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kpt fn run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kpt fn run local-resource/

The generated resources will appear in local-resource/

```
$ cat local-resource/*
...
```
