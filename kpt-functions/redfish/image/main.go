// Package main implements pod emulation function to run arbitrary scripts and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"fmt"
	"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish"
	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish/drivers/dell"
	"github.com/aodinokov/noctl-airship-poc/kpt-functions/redfish/drivers/dmtf"
)

func main() {
	log.Print("started")
	defer log.Print("finished")

	df := redfish.NewDriverFactory()
	drivers := []struct {
		vendor string
		model  string
		f      redfish.DriverConstructor
	}{
		{
			vendor: "dell",
			model:  "",
			f:      dell.NewDriver,
		},
		{
			vendor: "",
			model:  "",
			f:      dmtf.NewDriver,
		},
	}

	for _, driver := range drivers {
		if err := df.Register(driver.vendor, driver.model, driver.f); err != nil {
			fmt.Fprintf(os.Stderr, "Can't register driver for vendor %s, modle %s, err: %v\n",
				driver.vendor, driver.model, err)
			os.Exit(1)
		}
	}

	function := redfish.OperationFunction{DrvFactory: df}
	resourceList := &framework.ResourceList{FunctionConfig: &function.Config}

	cmd := framework.Command(resourceList, func() error {
		log.Print("entered")
		err := function.FinalizeInit(resourceList.Items)
		if err != nil {
			return err
		}
		log.Print("executing")
		return function.Execute()
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
