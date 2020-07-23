# Replacement function

This is an example of implementing a replacement function.
The feature set is similar to transformer [replacement transformer](https://github.com/kubernetes-sigs/kustomize/tree/master/plugin/someteam.example.com/v1/replacementtransformer)
but [MultiRef](https://github.com/aodinokov/noctl-airship-poc/blob/master/kpt-functions/replacement/function.go#L87) objects were added that allows to build strings based on several sources and put into several targets, similarly what [kpt-substitutions](https://github.com/kubernetes-sigs/kustomize/blob/master/kyaml/setters2/doc.go#L79) do.
To be compatible with Airship2 replacement plugin Regexp pattern feature was also added, even though MultiRef can overlap with this functionality.

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

The template is implemented as a go program, which reads a collection of input
Resource configuration, and looks for invalid configuration.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize config run local-resource/

This exits non-zero if there is an error.

## Running the Example

Run the function with:

    kustomize config run local-resource/

This will make the necessary changes in the yaml documents.
