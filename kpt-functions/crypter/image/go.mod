module github.com/aodinokov/noctl-airship-poc/kpt-functions/crypter/image

go 1.14

require (
	github.com/aodinokov/noctl-airship-poc/kpt-functions/crypter v0.0.0-20200824221209-9b7fb01e97a7
	github.com/spf13/cobra v1.0.0
	sigs.k8s.io/kustomize/kyaml v0.4.1
)

//replace github.com/aodinokov/noctl-airship-poc/kpt-functions/crypter => ../
