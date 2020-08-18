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
	for _, ref := range f.Config.Refs {
		node, err := ref.ObjRef.FindOne(items)
		if err != nil {
			return nil, err
		}

		key, err := Key(f.Config.Password)
		if err != nil {
			return nil, err
		}

		for _, fieldRef := range ref.FieldRefs {
			err := decryptField(node, fieldRef, key)
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

func decryptField(node *yaml.RNode, fieldRef string, key string) error {

	p, err := parseFieldRef(fieldRef)
	if err != nil {
		return err
	}
	/*
		s, err := node.String()
		if err != nil {
			return err
		}
		log.Printf("finding %v in\n%s", p, s)*/

	cn, err := node.Pipe(yaml.Lookup(p...))
	if err != nil {
		return err
	}

	if cn == nil {
		return fmt.Errorf("wasn't able to find %s", fieldRef)
	}

	if cn.YNode().Kind != yaml.ScalarNode {
		return fmt.Errorf("need scalar by path %s", fieldRef)
	}

	decyptedVal, err := Decrypt(yaml.GetValue(cn), key)
	if err != nil {
		return fmt.Errorf("wan't able to decrypt: %v", err)
	}

	err = cn.PipeE(yaml.FieldSetter{StringValue: decyptedVal})
	if err != nil {
		return fmt.Errorf("wan't able to set back: %v", err)
	}

	return nil
}
