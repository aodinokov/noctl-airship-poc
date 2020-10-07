module github.com/aodinokov/noctl-airship-poc/kpt-functions/templater/image

go 1.14

require (
	github.com/aodinokov/noctl-airship-poc/kpt-functions/templater v0.0.0-20201007163248-814fabfdc78f
	sigs.k8s.io/kustomize/kyaml v0.4.1
)

//replace github.com/aodinokov/noctl-airship-poc/kpt-functions/templater => ../
