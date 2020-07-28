This demo show the how [fn plugins](https://github.com/kubernetes-sigs/kustomize/tree/master/api/internal/plugins/fnplugin) work in kustomize 3.8.1
It's based on this [PR](https://review.opendev.org/#/c/724869/21) and the only changes that were made, added the following annotations:

```
  annotations:
    config.kubernetes.io/function: |
      container:
        image: quay.io/aodinokov/replacement-default:v0.0.1
```

in all documents that had kind ReplacementTransformer

```
t$ grep -Rni 'ReplacementTransformer'
manifests-example/composite/lma/replacements/prometheus.yaml:4:kind: ReplacementTransformer
manifests-example/composite/reference-cloud/replacements/replacements.yaml:4:kind: ReplacementTransformer
manifests-example/composite/openstack/replacements/keystone.yaml:4:kind: ReplacementTransformer
``` 

When kustomize tries to load plugin it now checks for annotation `config.kubernetes.io/function`.
If it present and can be parsed it assumes that the plugin configuration is kpt-function-configuration.
If not it fallbacks to the standard mechanisms of loading plugins. In another workd fnplugins take preference.

The description of functions can be found [here](https://googlecontainertools.github.io/kpt/concepts/functions/#functions-concepts).
Basically as [this](https://github.com/GoogleContainerTools/kpt/issues/646) issue states functions are very similar to the previous version of 
kustomize plugins, but it can be exec, container and starlark script.
For more implementation details please refer to [this](https://github.com/kubernetes-sigs/kustomize/pull/2597) and [this](https://github.com/kubernetes-sigs/kustomize/pull/2667).

How this demo works:
After modification of ReplacementTransformer configurations kustmoize calls docker container `quay.io/aodinokov/replacement-default:v0.0.1`
that accepts the same configuration and performs the same actions as ReplacementTransformer.
The source code for that function can be found [here](https://github.com/aodinokov/noctl-airship-poc/tree/master/kpt-functions/replacement).

Why it can be useful for Airship 2:
1. There was a big discussion how to deliver plugins so they would be included to airshipct. airshipct currently incorporates plugins and the chain of call is the following:
   airshipctl calls kustomize that calls another process of airshipctl that works as a kustomize plugin. There were several drawbacks for that approach, e.g. to validate
   Airship 2 manifests it's needed to install airshipctl (in another words they're kustomize manifests, but kustomize isn't enough to work with them). Another problem was
   that airshipctl wasn't possible to deliver in from of library. airhipctl binary has to be present on the host to be able to run manifests. 
   This implementation needs only docker installed, which anyway was a prerequisite for airshipctl. 
2. Kustomize community is going to [deprecate](https://github.com/GoogleContainerTools/kpt/issues/776) the previous form of plugins. That basically means that
   airshipctl needs to go to fn-plugins sooner or later. And as it's possible to see, it's better to do it now, because it's very easy at that point of time.
3. Fn-plugins give the greater flexibility level. With the previous approach we used in airshipctl it was necessary to include all plugins into airshipctl itself.
   That was an example of a pretty opiniated design. Now the kustomize modules developers may choose to use their own set of plugins for their modules.
   Moreover Airship 2 manifests repos may inlcude the source-code of the needed plugins to emphasize the manifests dependency from that functions.
   It's possible to separate manifests from airshipctl repos and use kustomize for writing some unit-tests for manifests. Build process should include build of 
   needed functions, downloading of the stable release of kustomize and running it with different values of VariableCatalogue.

To run just install docker and call ./run.sh. The script will download the neede version of kustomize and put the resulting resources to output.yaml.
During the first run it will be possible how docker pulls the needed function image `quay.io/aodinokov/replacement-default:v0.0.1`.
You can find the resulting output in the [expected_output.yaml](expected_output.yaml).
