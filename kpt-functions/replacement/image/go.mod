module github.com/aodinokov/noctl-airship-poc/kpt-functions/replacement/image

go 1.14

require (
	github.com/aodinokov/noctl-airship-poc/kpt-functions/replacement v0.0.1
	sigs.k8s.io/kustomize/kyaml v0.4.1
)

replace github.com/aodinokov/noctl-airship-poc/kpt-functions/replacement => ../
