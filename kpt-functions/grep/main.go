package main

import (
	"fmt"
	"log"
	"os"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/grep/function"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
)

func main() {
	cfg := function.Config{}
	resourceList := &framework.ResourceList{FunctionConfig: &cfg}

	cmd := framework.Command(resourceList, func() error {
		fn, err := function.NewFilter(&cfg)
		if err != nil {
			log.Printf("function creation failed: %v", err)
			return err
		}

		items, err := fn.Filter(resourceList.Items)
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
