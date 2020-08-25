package crypter

import (
	"fmt"
	//"log"
	"os"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/replacement"
)

type Ref struct {
	ObjRef    *replacement.SourceObjRef `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string                  `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

type FunctionConfig struct {
	OldPassword string `json:"oldPassword,omitempty" yaml:"oldPassword,omitempty"`
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	DryRun   bool   `json:"dryRun,omitempty" yaml:"dryRun,omitempty"`
	Operation string `json:"operation,omitempty" yaml:"operation,omitempty"`
	Refs     []Ref  `json:"refs,omitempty" yaml:"refs,omitempty"`
}

type Function struct {
	Config *FunctionConfig
	OldKey string
	Key string
}

func NewFunction(cfg *FunctionConfig) (*Function, error) {
	// read env var and override the values from config
	val, ok := os.LookupEnv("crypter_password")
	if ok {
		cfg.Password = val
	}

	val, ok = os.LookupEnv("crypter_old_password")
        if ok {
                cfg.OldPassword = val
        }

	val, ok = os.LookupEnv("crypter_dryrun")
	if ok {
		cfg.DryRun = true
	}

	fn := Function{Config: cfg}
	return &fn, nil
}

func (f *Function) Exec(items []*yaml.RNode) ([]*yaml.RNode, error) {
	if f.Config.Operation == "" {
		f.Config.Operation = "decrypt"
	}

	if f.Config.Password != "" {
		key, err := Key(f.Config.Password)
		if err != nil {
			return nil, err
		}
		f.Key = key
	}
	if f.Config.OldPassword != "" {
		key, err := Key(f.Config.OldPassword)
		if err != nil {
			return nil, err
		}
		f.OldKey = key
	}

	for _, ref := range f.Config.Refs {
		nodes, err := ref.ObjRef.Filter(items)
		if err != nil {
			return nil, err
		}

		for _, node := range nodes {
			for _, fieldRef := range ref.FieldRefs {
				err := f.execFieldOp(node, fieldRef)
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return items, nil
}

func (f *Function) execFieldOp(node *yaml.RNode, fieldRef string) error {

	p, err := replacement.ParseFieldRef(fieldRef)
	if err != nil {
		return err
	}

	return f.execFieldOpImpl(node, fieldRef, p)
}

func (f *Function) execFieldOpImpl(node *yaml.RNode, fieldRef string, fieldRefs []string) error {
	for i := range fieldRefs {
		// handle case with * to scan all elements
		if fieldRefs[i] == "*" {
			cn, err := node, error(nil)
			if i > 0 {
				//log.Printf("%d %v %v", i, fieldRefs, fieldRefs[:i])
				cn, err = node.Pipe(yaml.Lookup(fieldRefs[:i]...))
				if err != nil {
					return err
				}
			}

			if cn.YNode().Kind != yaml.SequenceNode {
				return fmt.Errorf("asterisk is applicable only for SequenceNode")
			}

			for _, n := range cn.Content() {
				//log.Printf("%v %v", fieldRefs, fieldRefs[i+1:])
				err := f.execFieldOpImpl(yaml.NewRNode(n), fieldRef, fieldRefs[i+1:])
				if err != nil {
					return err
				}
			}
			return nil

		}
	}

	cn, err := node.Pipe(yaml.Lookup(fieldRefs...))
	if err != nil {
		return err
	}

	if cn == nil {
		return fmt.Errorf("wasn't able to find %s", fieldRef)
	}

	if cn.YNode().Kind != yaml.ScalarNode {
		return fmt.Errorf("need scalar by path %s", fieldRef)
	}

	resultingVal := fmt.Sprintf("performed %s for %s", f.Config.Operation, fieldRef)
	if !f.Config.DryRun {
		switch(f.Config.Operation) {
		case "decrypt":
			resultingVal, err = Decrypt(yaml.GetValue(cn), f.Key)
			if err != nil {
				return fmt.Errorf("wan't able to decrypt: %v", err)
			}
		case "encrypt":
			resultingVal, err = Encrypt(yaml.GetValue(cn), f.Key)
			if err != nil {
				return fmt.Errorf("wan't able to encrypt: %v", err)
			}
		case "rotate":
			resultingVal, err = Decrypt(yaml.GetValue(cn), f.OldKey)
			if err != nil {
				return fmt.Errorf("wan't able to decrypt: %v", err)
			}
			resultingVal, err = Encrypt(resultingVal, f.Key)
			if err != nil {
				return fmt.Errorf("wan't able to encrypt: %v", err)
			}
		}
	}

	err = cn.PipeE(yaml.FieldSetter{StringValue: resultingVal})
	if err != nil {
		return fmt.Errorf("wan't able to set back: %v", err)
	}

	return nil

}
