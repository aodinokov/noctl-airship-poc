// Package main implements pod emulation function to run arbitrary scripts and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish"
)

func main() {
	log.Print("started")
	defer log.Print("Finished")

	df := redfish.NewDriverFactory()
	if err := df.Register("", "", nil/*TODO: put here default/default driver fn */); err != nil {
		fmt.Fprintf(os.Stderr, "Can't register default driver\n")
		os.Exit(1)
	}

	function := redfish.OperationFunction{DrvFactory: df}
	resourceList := &framework.ResourceList{FunctionConfig: &function.Config}

	cmd := framework.Command(resourceList, func() error {
		err := function.FinalizeInit(resourceList.Items)
		if err != nil {
			return err
		}
		return function.Execute()
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
