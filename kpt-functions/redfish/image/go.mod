module github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish/image/

go 1.14

require (
	github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish v0.0.0-20200707202325-6e8ad2f9a677
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	sigs.k8s.io/kustomize/kyaml v0.1.11
)

replace github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish v0.0.0-20200707202325-6e8ad2f9a677 => ../
