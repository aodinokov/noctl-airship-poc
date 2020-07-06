package redfish

import (
	"fmt"

	k8sv1 "k8s.io/api/core/v1"
	metal3v1alpha1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"

	"sigs.k8s.io/kustomize/kyaml/yaml"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
)

type Operation struct {
	Action string `yaml:"action"`
	Args []string `yaml:"args,omitempty"`
}

type RedfishOperationFunctionConfig struct {
	Spec struct {
		Operations []Operation `yaml:"operations,omitempty"`
		BmhRef struct {
			Name string `yaml:"name"`
			Namespace string `yaml:"namespace"`
		} `yaml:"bmhRef"`
	} `yaml:"spec,omitempty"`
}

type RedfishOperationFunction struct {
	DrvFactory *DriverFactory

	Config RedfishOperationFunctionConfig

	Bmh *metal3v1alpha1.BareMetalHost
	CredentialsSecret *k8sv1.Secret

	Items []*yaml.RNode

	Drv Driver
}

// Check if the read values are valid
// Perform some caching initialization
func (f *RedfishOperationFunction) FinalizeInit(items []*yaml.RNode) error {
	if f.DrvFactory == nil {
		return fmt.Errorf("driver factory isn't initialized")
	}

	f.Items = items

	if err := f.findAndKeepBmh(); err != nil {
		return err
	}
	if err := f.findAndKeepCredentialsSecret(); err != nil {
		return err
	}

	// TODO: make some invariant check

	fn, err := f.DrvFactory.GetCreateDriverFn(f.Bmh.Spec.RootDeviceHints.Vendor, f.Bmh.Spec.RootDeviceHints.Model)
	if err != nil {
		return err
	}
	drv, err := fn(f)
	if err != nil {
		return err
	}
	f.Drv = drv

	return nil
}

func (f *RedfishOperationFunction) findAndKeepBmh() error {
	c := complexFilter{
		Filters: []kio.Filter{
			filters.GrepFilter{Path: []string{"apiVersion"}, Value: "metal3.io/v1alpha1", },
			filters.GrepFilter{Path: []string{"kind"}, Value: "BareMetalHost", },
			filters.GrepFilter{Path: []string{"metadata", "name"}, Value: f.Config.Spec.BmhRef.Name, },
			filters.GrepFilter{Path: []string{"metadata", "namespace"}, Value: f.Config.Spec.BmhRef.Namespace, },
		},
	}
	nodes, err := c.Filter(f.Items)
	if err != nil {
		return err
	}
	if len(nodes) != 1 {
		return fmt.Errorf("looked for BareMetalHost:metal3.io/v1alpha1 with name %s, namespace %s, expected 1, found %d",
			f.Config.Spec.BmhRef.Name,
			f.Config.Spec.BmhRef.Namespace,
			len(nodes))
	}
	// Convert to BMH struct
	b, err := yaml.Marshal(nodes[0])
	if err != nil {
		return err
	}

	f.Bmh = &metal3v1alpha1.BareMetalHost{}
	err = yaml.Unmarshal(b, f.Bmh)
	if err != nil {
		return err
	}
	return nil
}

func (f *RedfishOperationFunction) findAndKeepCredentialsSecret() error {
	c := complexFilter{
		Filters: []kio.Filter{
			filters.GrepFilter{Path: []string{"apiVersion"}, Value: "v1", },
			filters.GrepFilter{Path: []string{"kind"}, Value: "Secret", },
			filters.GrepFilter{Path: []string{"metadata", "name"}, Value: f.Bmh.Spec.BMC.CredentialsName, },
			filters.GrepFilter{Path: []string{"metadata", "namespace"}, Value: f.Config.Spec.BmhRef.Namespace, },

		},
	}
	nodes, err := c.Filter(f.Items)
	if err != nil {
		return err
	}
	if len(nodes) != 1 {
		return fmt.Errorf("looked for Secret:v1 with name %s, namespace %s, expected 1, found %d",
			f.Bmh.Spec.BMC.CredentialsName,
			f.Config.Spec.BmhRef.Namespace,
			len(nodes))
	}
	// Convert to Secret struct
	b, err := yaml.Marshal(nodes[0])
	if err != nil {
		return err
	}
	f.CredentialsSecret = &k8sv1.Secret{}
	err = f.CredentialsSecret.Unmarshal(b)
	if err != nil {
		return err
	}
	return nil
}

func (f *RedfishOperationFunction) Execute() error {
	for i := range f.Config.Spec.Operations {
		if err := f.execOperation(i); err != nil {
			return err
		}
	}
	return nil
}

func (f *RedfishOperationFunction) execOperation(i int) error {
	// TODO: just call the driver
	return nil
}

