package crypter

import (
	"bytes"
	"fmt"
	//"log"
	"os"
	"unicode/utf8"

	"sigs.k8s.io/kustomize/kyaml/yaml"

	"github.com/aodinokov/noctl-airship-poc/kpt-functions/replacement"
)

type Ref struct {
	ObjRef    *replacement.SourceObjRef `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string                  `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

type FunctionConfig struct {
	Password string `json:"password,omitempty" yaml:"password,omitempty"`
	DryRun   bool   `json:"dryRun,omitempty" yaml:"dryRun,omitempty"`
	Refs     []Ref  `json:"refs,omitempty" yaml:"refs,omitempty"`
}

type Function struct {
	Config *FunctionConfig
}

func NewFunction(cfg *FunctionConfig) (*Function, error) {
	// read env var and override the values from config
	val, ok := os.LookupEnv("crypter_password")
	if ok {
		cfg.Password = val
	}

	val, ok = os.LookupEnv("crypter_dryrun")
	if ok {
		cfg.DryRun = true
	}

	fn := Function{Config: cfg}
	return &fn, nil
}

func (f *Function) Exec(items []*yaml.RNode) ([]*yaml.RNode, error) {
	key, err := Key(f.Config.Password)
	if err != nil {
		return nil, err
	}

	for _, ref := range f.Config.Refs {
		node, err := ref.ObjRef.FindOne(items)
		if err != nil {
			return nil, err
		}

		for _, fieldRef := range ref.FieldRefs {
			err := f.decryptField(node, fieldRef, key)
			if err != nil {
				return nil, err
			}
		}
	}
	return items, nil
}

// TODO: to get from replacement?
func parseFieldRef(in string) ([]string, error) {
	var cur bytes.Buffer
	out := []string{}
	var state int
	for i := 0; i < len(in); {
		r, size := utf8.DecodeRuneInString(in[i:])

		switch state {
		case 0: // initial state
			if r == '.' {
				if cur.String() != "" {
					out = append(out, cur.String())
					cur = bytes.Buffer{}
				}
			} else if r == '[' {
				if cur.String() != "" {
					out = append(out, cur.String())
					cur = bytes.Buffer{}
				}
				cur.WriteRune(r)
				state = 1
			} else {
				cur.WriteRune(r)
			}
		case 1: // state inside []
			cur.WriteRune(r)
			if r == ']' {
				state = 0
			}
		}
		i += size
	}

	if state != 0 {
		return nil, fmt.Errorf("unclosed [")
	}

	return append(out, cur.String()), nil
}

func (f *Function) decryptField(node *yaml.RNode, fieldRef string, key string) error {

	p, err := parseFieldRef(fieldRef)
	if err != nil {
		return err
	}

	return f.decryptFieldImpl(node, fieldRef, p, key)
}

func (f *Function) decryptFieldImpl(node *yaml.RNode, fieldRef string, fieldRefs []string, key string) error {
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
				err := f.decryptFieldImpl(yaml.NewRNode(n), fieldRef, fieldRefs[i+1:], key)
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

	decyptedVal := fmt.Sprintf("decrypted %s", fieldRef)
	if !f.Config.DryRun {
		decyptedVal, err = Decrypt(yaml.GetValue(cn), key)
		if err != nil {
			return fmt.Errorf("wan't able to decrypt: %v", err)
		}
	}

	err = cn.PipeE(yaml.FieldSetter{StringValue: decyptedVal})
	if err != nil {
		return fmt.Errorf("wan't able to set back: %v", err)
	}

	return nil

}
