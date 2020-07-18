package replacement

import (
	"fmt"

	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

type FunctionConfig struct {
	Replacements []types.Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
}

type Function struct {
	Config *FunctionConfig
}

func NewFunction(cfg *FunctionConfig) (*Function, error) {
	for _, r := range cfg.Replacements {
		if r.Source == nil {
			return nil, fmt.Errorf("`from` must be specified in one replacement")
		}
		if r.Target == nil {
			return nil, fmt.Errorf("`to` must be specified in one replacement")
		}
		count := 0
		if r.Source.ObjRef != nil {
			count += 1
		}
		if r.Source.Value != "" {
			count += 1
		}
		if count > 1 {
			return nil, fmt.Errorf("only one of fieldref and value is allowed in one replacement")
		}
	}

	fn := Function{Config: cfg}
	return &fn, nil
}

func (f *Function) Exec(items []*yaml.RNode) error {
	for _, r := range f.Config.Replacements {
		var replacement interface{}
		if r.Source.ObjRef != nil {
			// TODO: replacement, err = getReplacement(m, r.Source.ObjRef, r.Source.FieldRef)
			//if err != nil {
			//	return err
			//}
		}
		if r.Source.Value != "" {
			replacement = r.Source.Value
		}

		fmt.Printf("The replacement is %s\n", replacement)

		//err = substitute(m, r.Target, replacement)
		//if err != nil {
		//	return err
		//}
	}
	return nil
}


