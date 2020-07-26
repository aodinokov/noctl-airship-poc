This demo shows that kustominze actually has ability to
remove needed resources using SMP construction:

```
$patch: delete
```

helloWorld/arbitrary.yaml contains airship-specific resource.
./run.sh script runs kustomize 3.8.1 2 times:
without patch
and with SMP patch helloWorld/arbitraryCleanup.yaml
and shows the difference.
When you run it it's clean that the resource disappears in the second case:

```
$ ./run.sh
65,75d64
< ---
< apiVersion: airshipit.org/v1alpha1
< data:
<   altGreeting: some data
<   enableRisky: more data
< kind: VariableCatalogue
< metadata:
<   labels:
<     airshipit.org/deploy-k8s: "false"
<     app: hello
<   name: openstack-endpoint-catalogue
```
