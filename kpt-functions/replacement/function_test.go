package replacement

import (
	"testing"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestGetFieldValue(t *testing.T) {
	ts := []struct {
		InYaml        string
		InField       string
		ExpectedError bool
		ExpectedVal   string
	}{
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField:     "a.b.c",
			ExpectedVal: "value",
		},
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField:     ".a.b.c",
			ExpectedVal: "value",
		},
		{
			InYaml: `
a:
  b: |
    c:
      d: innerValue
`,
			InField:     ".a.b|c.d",
			ExpectedVal: "innerValue",
		},
		{
			InYaml: `
a:
  b: |
    c:
      d: innerValue1
      e: "f: innerValue2"
`,
			InField:     "a.b|c.e|f",
			ExpectedVal: "innerValue2",
		},
	}

	for _, ti := range ts {
		node, err := yaml.Parse(ti.InYaml)
		if err != nil {
			t.Errorf("wasn't able to parse inYaml %s: %v, trying to continue", ti.InYaml, err)
			continue
		}

		val, err := getFieldValue(node, ti.InField)
		if err != nil && ti.ExpectedError {
			continue
		}
		if err != nil {
			t.Errorf("didn't expect error for field: %s yaml %s: %v", ti.InField, ti.InYaml, err)
		}
		if val != ti.ExpectedVal {
			t.Errorf("unexpected value %s for field: %s yaml %s. Expected %s",
				val, ti.InField, ti.InYaml, ti.ExpectedVal)
		}
	}

}

func TestSetFieldValue(t *testing.T) {
	ts := []struct {
		InYaml        string
		InField       string
		InValue       string
		ExpectedError bool
		ExpectedYaml  string
	}{
		{
			InYaml: `
a:
  b:
    c: value
`,
			InField: "a.b.c",
			InValue: "newvalue",
			ExpectedYaml: `
a:
  b:
    c: newvalue
`,
		},
	}

	for _, ti := range ts {
		node, err := yaml.Parse(ti.InYaml)
		if err != nil {
			t.Errorf("wasn't able to parse inYaml %s: %v, trying to continue", ti.InYaml, err)
			continue
		}
		err = setFieldValue(node, ti.InField, ti.InValue)
		if err != nil && ti.ExpectedError {
			continue
		}
		if err != nil {
			t.Errorf("didn't expect error for field: %s yaml %s: %v", ti.InField, ti.InYaml, err)
		}
		resYaml, err := node.String()
		if err != nil {
			t.Errorf("got unexpected error converting node back for inYaml %s: %v", ti.InYaml, err)
			continue
		}
		if resYaml != ti.ExpectedYaml[1:] {
			t.Errorf("expected \n%s, got \n%s", ti.ExpectedYaml, resYaml)
		}
	}
}
