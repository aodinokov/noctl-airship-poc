# Redfish ISO started

This is an example implements a function that reads the metal3 BareMetalHost
related configuration from ResourceList.items and performs operations,
listed in the list of operations in the function configuration.

This example is written in `go` and uses the `kyaml` libraries for parsing the
input and writing the output.  Writing in `go` is not a requirement.

## Function implementation

The function is implemented as an [image](image), and built using `make image`.

The template is implemented as a go program, which reads a collection of input
Resource configuration, and looks for invalid configuration.

## Function invocation

The function is invoked by authoring a [local Resource](local-resource)
with `metadata.annotations.[config.kubernetes.io/function]` and running:

    kustomize fn run local-resource/

This exits non-zero if there is an error.

## Running the Example

Exec mode:
Run the validator with:
    
    make # will create binary. Make sure that the correct path to it is set 
         # in the local-resource/example-use.yaml 

    kustomize fn run local-resource/ --enable-exec

Container mode
Run the validator with:

    make image                         # will generate the image with tag
                                       # quay.io/airshipit/kpt-functions/redfish-debian_stable:v0.0.1
    vi local-resource/example-use.yaml # change exec.path to container.image: quay.io/airshipit/kpt-functions/redfish-debian_stable:v0.0.1
                                       # change BMC and image URLs from localhost to something that can be reached from docker container

    kustomize fn run local-resource/ --network

This will send a series of redfish commands to boot the system from the iso
provided in the configuration.
