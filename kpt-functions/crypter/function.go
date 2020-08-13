package crypter

import (
	//"bytes"
	"fmt"

	//"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/replacement"
)

type Ref struct {
	ObjRef    *replacement.SourceObjRef `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string                  `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

type FunctionConfig struct {
	Refs []Ref `json:"refs,omitempty" yaml:"refs,omitempty"`
}

type Function struct {
	Config *FunctionConfig

	key string
}

func NewFunction(cfg *FunctionConfig) (*Function, error) {
	// read env var for password

	fn := Function{Config: cfg}
	return &fn, nil
}

func (f *Function) Exec(items []*yaml.RNode) ([]*yaml.RNode, error) {
	for _, ref := range f.Config.Refs {
		node, err := ref.ObjRef.FindOne(items)
		if err != nil {
			return nil, err
		}

		for _, fieldRef := range ref.FieldRefs {
			err := decryptField(node, fieldRef, f.key)
			if err != nil {
				return nil, err
			}
		}
	}
	return items, nil
}

func decryptField(node *yaml.RNode, fieldRef string, key string) error {
	cn, err := node.Pipe(yaml.Lookup(fieldRef))
	if err != nil {
		return err
	}

	if cn.YNode().Kind != yaml.ScalarNode {
		return fmt.Errorf("need scalar by path %s", fieldRef)
	}

	decyptedVal, err := decrypt(yaml.GetValue(cn), key)
	if err != nil {
		return fmt.Errorf("wan't able to decrypt: %v", err)
	}

	err = cn.PipeE(yaml.FieldSetter{StringValue: decyptedVal})
	if err != nil {
		return fmt.Errorf("wan't able to set back: %v", err)
	}

	return nil
}
