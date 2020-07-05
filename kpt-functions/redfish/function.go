package redfish

import (
	"fmt"

	k8sv1 "k8s.io/api/core/v1"
	metal3v1alpha1 "github.com/metal3-io/baremetal-operator/pkg/apis/metal3/v1alpha1"

	"sigs.k8s.io/kustomize/kyaml/yaml"
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
	Config RedfishOperationFunctionConfig

	Bmh *metal3v1alpha1.BareMetalHost
	CredentialsSecret *k8sv1.Secret

	Items []*yaml.RNode
}

// Check if the read values are valid
// Perform some caching initialization
func (f *RedfishOperationFunction) FinalizeInit(items []*yaml.RNode) error {
	f.Items = items

	if err := f.findAndKeepBmh(); err != nil {
		return err
	}
	if err := f.findAndKeepCredentialsSecret(); err != nil {
		return err
	}

	// TODO: make some invariant check

	return nil
}

func (f *RedfishOperationFunction) findAndKeepBmh() error {
	// TODO: check if kyaml has a filter to look by meta fields
	for _, r := range f.Items {
		meta, err := r.GetMeta()
		if err != nil {
			return err
		}

		if meta.Kind != "BareMetalHost" ||  meta.Name != f.Config.Spec.BmhRef.Name || meta.Namespace != f.Config.Spec.BmhRef.Namespace {
			continue
		}

		if meta.APIVersion != "metal3.io/v1alpha1" {
			return fmt.Errorf("unsupported BareMetalHost %s apiVerion: %s", 
				meta.Name, meta.APIVersion)
		}
		// Convert to BMH struct
		b, err := yaml.Marshal(r)
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

	return fmt.Errorf("wasn't able to find BareMetalHost %s in namespace %s",
		f.Config.Spec.BmhRef.Name,
		f.Config.Spec.BmhRef.Namespace)
}

func (f *RedfishOperationFunction) findAndKeepCredentialsSecret() error {
	// TODO: see above
	for _, r := range f.Items {
		meta, err := r.GetMeta()
		if err != nil {
			return err
		}

		if meta.Kind != "Secret" || meta.Name != f.Bmh.Spec.BMC.CredentialsName || meta.Namespace != f.Config.Spec.BmhRef.Namespace {
			continue
		}

		if meta.APIVersion != "v1" {
			return fmt.Errorf("unsupported Secret %s apiVerion: %s",
				meta.Name, meta.APIVersion)
		}
		// Convert to Secret struct
		b, err := yaml.Marshal(r)
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
	return fmt.Errorf("wasn't able to find BareMetalHost %s in namespace %s",
		f.Bmh.Spec.BMC.CredentialsName,
		f.Config.Spec.BmhRef.Namespace)
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
	// TODO: init context 
	return nil
}

