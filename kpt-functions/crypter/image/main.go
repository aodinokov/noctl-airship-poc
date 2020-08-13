// Package main implements pod emulation function to run arbitrary scripts and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/crypter"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

func main() {
	log.Print("started")
	defer log.Print("Finished")

	cfg := crypter.FunctionConfig{}
	resourceList := &framework.ResourceList{FunctionConfig: &cfg}

	cmd := framework.Command(resourceList, func() error {
		fn, err := crypter.NewFunction(&cfg)
		if err != nil {
			log.Printf("function creation failed: %v", err)
			return err
		}

		items, err := fn.Exec(resourceList.Items)
		if err != nil {
			return err
		}
		resourceList.Items = items
		return nil
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
