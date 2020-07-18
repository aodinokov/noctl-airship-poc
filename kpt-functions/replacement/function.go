package replacement

import (
	"fmt"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Gvk identifies a Kubernetes API type.
// https://github.com/kubernetes/community/blob/master/contributors/design-proposals/api-machinery/api-group.md
type Gvk struct {
	Group   string `json:"group,omitempty" yaml:"group,omitempty"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	Kind    string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

// Selector specifies a set of resources.
// Any resource that matches intersection of all conditions
// is included in this set.
type Selector struct {
	Gvk       `json:",inline,omitempty" yaml:",inline,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`

	// AnnotationSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource annotations.
	AnnotationSelector string `json:"annotationSelector,omitempty" yaml:"annotationSelector,omitempty"`

	// LabelSelector is a string that follows the label selection expression
	// https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#api
	// It matches with the resource labels.
	LabelSelector string `json:"labelSelector,omitempty" yaml:"labelSelector,omitempty"`
}

// Target refers to a kubernetes object by Group, Version, Kind and Name
// gvk.Gvk contains Group, Version and Kind
// APIVersion is added to keep the backward compatibility of using ObjectReference
// for Var.ObjRef
type SourceObjRef struct {
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Gvk        `json:",inline,omitempty" yaml:",inline,omitempty"`
	Name       string `json:"name" yaml:"name"`
	Namespace  string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// Source defines where a substitution is from
// It can from two different kinds of sources
//  - from a field of one resource
//  - from a several sources and go template
//  - from a string
type Source struct {
	ObjRef   *SourceObjRef `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRef string        `json:"fieldref,omitempty" yaml:"fiedldref,omitempty"`

	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	Multiref *struct {
		Sources []struct {
			ObjRef   *SourceObjRef `json:"objref" yaml:"objref"`
			FieldRef string        `json:"fieldref" yaml:"fiedldref"`
		}
		Template string `json:"template" yaml:"template"`
	} `json:"multiref,omitempty" yaml:"multiref,omitempty"`
}

// ReplTarget defines where a substitution is to.
type Target struct {
	ObjRef    *Selector `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRefs []string  `json:"fieldrefs,omitempty" yaml:"fieldrefs,omitempty"`
}

type Replacement struct {
	Source *Source `json:"source" yaml:"source"`
	Target *Target `json:"target" yaml:"target"`
}

type FunctionConfig struct {
	Replacements []Replacement `json:"replacements,omitempty" yaml:"replacements,omitempty"`
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
		if r.Source.Multiref != nil {
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
		source, err := prepareSource(items, r.Source)
		if err != nil {
			return err
		}

		fmt.Printf("The source is %s\n", source)

		err = apply(items, source, r.Target)
		if err != nil {
			return err
		}
	}
	return nil
}

func prepareSource(items []*yaml.RNode, s *Source) (string, error) {
//	if 
	return "", nil
}

func apply(items []*yaml.RNode, source string, t *Target) error {
	return nil
}
