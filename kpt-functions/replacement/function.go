package replacement

import (
	"bytes"
	"fmt"
	"regexp"
	"text/template"

	"github.com/Masterminds/sprig"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/filters"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

var (
	// substring substitutions are appended to paths as: ...%VARNAME%
	substringPatternRegex = regexp.MustCompile(`(\S+)%(\S+)%$`)
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

type MultiSourceObjRef struct {
	Refs []struct {
		ObjRef   *SourceObjRef `json:"objref" yaml:"objref"`
		FieldRef string        `json:"fieldref" yaml:"fieldref"`
	} `json:"refs,omitempty" yaml:"refs,omitempty"`
	Template string `json:"template" yaml:"template"`
}

// Source defines where a substitution is from
// It can from two different kinds of sources
//  - from a field of one resource
//  - from a several sources and go template
//  - from a string
type Source struct {
	Value string `json:"value,omitempty" yaml:"value,omitempty"`

	ObjRef   *SourceObjRef `json:"objref,omitempty" yaml:"objref,omitempty"`
	FieldRef string        `json:"fieldref,omitempty" yaml:"fieldref,omitempty"`

	MultiRef *MultiSourceObjRef `json:"multiref,omitempty" yaml:"multiref,omitempty"`
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

func (r Replacement) Validate() error {
	if r.Source == nil {
		return fmt.Errorf("`from` must be specified in one replacement")
	}
	if r.Target == nil {
		return fmt.Errorf("`to` must be specified in one replacement")
	}
	count := 0
	if r.Source.ObjRef != nil {
		count += 1
	}
	if r.Source.Value != "" {
		count += 1
	}
	if r.Source.MultiRef != nil {
		count += 1
	}
	// @aodinokov, Should it be equal 1?
	if count > 1 {
		return fmt.Errorf("only one of fieldref and value is allowed in one replacement")
	}
	return nil
}

func (cfg FunctionConfig) Validate() error {
	for _, r := range cfg.Replacements {
		if err := r.Validate(); err != nil {
			return err
		}
	}
	return nil
}

func NewFunction(cfg FunctionConfig) (*Function, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Function{Config: &cfg}, nil
}

func (objRef SourceObjRef) prepareValue(items []*yaml.RNode, fieldRef string) (interface{}, error) {
	node, err := objRef.FindOne(items)
	if err != nil {
		return "", err
	}
	if fieldRef == "" {
		fieldRef = ".metadata.name"
	}
	return getFieldValue(node, fieldRef)
}

func (mObjRef MultiSourceObjRef) prepareValue(items []*yaml.RNode) (string, error) {
	data := struct {
		Values []string
	}{
		Values: make([]string, 0, len(mObjRef.Refs)),
	}
	for i := range mObjRef.Refs {
		v, err := mObjRef.Refs[i].ObjRef.prepareValue(items, mObjRef.Refs[i].FieldRef)
		if err != nil {
			return "", fmt.Errorf("error preparing multiref %v ref %d: %w", mObjRef, i, err)
		}
		sv, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("multiref sources can be scalars only")
		}
		data.Values = append(data.Values, sv)
	}

	var out bytes.Buffer
	tmpl, err := template.New("tmpl").Funcs(sprig.TxtFuncMap()).Parse(mObjRef.Template)
	if err != nil {
		return "", fmt.Errorf("error parsing template %s: %w", mObjRef.Template, err)
	}

	if err := tmpl.Execute(&out, data); err != nil {
		return "", fmt.Errorf("error executing template %s, data %v: %w",
			mObjRef.Template, data, err)
	}

	return out.String(), nil
}

func (s Source) prepareValue(items []*yaml.RNode) (interface{}, error) {
	if s.Value != "" {
		return s.Value, nil
	}
	if s.ObjRef != nil {
		return s.ObjRef.prepareValue(items, s.FieldRef)
	}
	if s.MultiRef != nil {
		return s.MultiRef.prepareValue(items)
	}
	// @aodinokov, Should it be an error?
	return "", nil
}

func (f Function) Exec(items []*yaml.RNode) error {
	for _, r := range f.Config.Replacements {
		value, err := r.Source.prepareValue(items)
		if err != nil {
			return err
		}

		if err := apply(items, r.Target, value); err != nil {
			return err
		}
	}

	return nil
}

func apply(items []*yaml.RNode, t *Target, value interface{}) error {
	matching, err := t.ObjRef.Filter(items)
	if err != nil {
		return fmt.Errorf("error filtering by objref %v: %w", t.ObjRef, err)
	}

	for _, node := range matching {
		for _, fieldref := range t.FieldRefs {
			if err := setFieldValueHandlingRegex(node, fieldref, value); err != nil {
				return fmt.Errorf("error setting value for objref %v, fieldref %s, value %s, node %v: %w",
					t.ObjRef, fieldref, value, node, err)
			}
		}
	}
	return nil
}

func (g *Gvk) Filters() ([]kio.Filter, error) {
	flts := []kio.Filter{}
	if g.Group != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"group"}, Value: g.Group})
	}
	if g.Version != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"version"}, Value: g.Version})
	}
	if g.Kind != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"kind"}, Value: g.Kind})
	}
	return flts, nil
}

func (s *Selector) Filters() ([]kio.Filter, error) {
	flts, err := s.Gvk.Filters()
	if err != nil {
		return nil, err
	}
	if s.Name != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "name"}, Value: s.Name})
	}
	if s.Namespace != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "namespace"}, Value: s.Namespace})
	}
	if s.AnnotationSelector != "" {
		flts = append(flts, LabelFilter{Path: []string{"metadata", "annotations"}, Selector: s.AnnotationSelector})
	}
	if s.LabelSelector != "" {
		flts = append(flts, LabelFilter{Path: []string{"metadata", "labels"}, Selector: s.LabelSelector})
	}
	return flts, nil
}

func (s *Selector) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	flts, err := s.Filters()
	if err != nil {
		return nil, err
	}
	return findMatching(items, flts)
}

func (s *SourceObjRef) Filters() ([]kio.Filter, error) {
	flts := []kio.Filter{}
	if s.APIVersion != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"apiVersion"}, Value: s.APIVersion})
	}

	gflts, err := s.Gvk.Filters()
	if err != nil {
		return nil, err
	}
	flts = append(flts, gflts...)

	if s.Name != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "name"}, Value: s.Name})
	}
	if s.Namespace != "" {
		flts = append(flts, filters.GrepFilter{Path: []string{"metadata", "namespace"}, Value: s.Namespace})
	}
	return flts, nil
}

func (s *SourceObjRef) Filter(items []*yaml.RNode) ([]*yaml.RNode, error) {
	flts, err := s.Filters()
	if err != nil {
		return nil, err
	}

	return findMatching(items, flts)
}

func (s *SourceObjRef) FindOne(items []*yaml.RNode) (*yaml.RNode, error) {
	matching, err := s.Filter(items)
	if err != nil {
		return nil, err
	}

	if len(matching) > 1 {
		return nil, fmt.Errorf("found more than one resources matching from %v", s)
	}

	if len(matching) == 0 {
		return nil, fmt.Errorf("failed to find one resource matching from %v", s)
	}

	return matching[0], nil
}

func findMatching(items []*yaml.RNode, flts []kio.Filter) ([]*yaml.RNode, error) {
	p := kio.Pipeline{
		Inputs:  []kio.Reader{&kio.PackageBuffer{Nodes: items}},
		Filters: flts,
		Outputs: []kio.Writer{&kio.PackageBuffer{}},
	}

	if err := p.Execute(); err != nil {
		return nil, err
	}
	return p.Outputs[0].(*kio.PackageBuffer).Nodes, nil
}

func setFieldValueHandlingRegex(node *yaml.RNode, fieldRef string, value interface{}) error {
	// fieldref can contain substtringpattern for regexp - we need to get it
	substringPattern := ""
	groups := substringPatternRegex.FindStringSubmatch(fieldRef)
	if len(groups) == 3 {
		fieldRef = groups[1]
		substringPattern = groups[2]

		svalue, ok := value.(string)
		if !ok {
			return fmt.Errorf("regex fieldref can be used only with scalar sources")
		}

		// calculate real value
		v, err := getFieldValue(node, fieldRef)
		if err != nil {
			return fmt.Errorf("wasn't able to get value for node %v, fieldref %s",
				node, fieldRef)
		}

		sv, ok := v.(string)
		if !ok {
			return fmt.Errorf("regex fieldref can be used only for scalar fields target: %s", fieldRef)
		}

		p := regexp.MustCompile(substringPattern)
		if !p.MatchString(sv) {
			return fmt.Errorf("wasn't able to match pattern %s with value %s",
				substringPattern, sv)
		}
		value = p.ReplaceAllString(sv, svalue)
	}

	return setFieldValue(node, fieldRef, value)
}
