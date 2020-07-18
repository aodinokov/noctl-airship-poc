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
			t.Errorf("wasn't able to parse inYaml %s, trying to continue", ti.InYaml)
			continue
		}

		val, err := getFieldValue(node, ti.InField)
		if err != nil && ti.ExpectedError {
			continue
		}
		if err != nil {
			t.Errorf("didn't expect error for field: %s yaml %s", ti.InField, ti.InYaml)
		}
		if val != ti.ExpectedVal {
			t.Errorf("unexpected value %s for field: %s yaml %s. Expected %s",
				val, ti.InField, ti.InYaml, ti.ExpectedVal)
		}
	}

}
