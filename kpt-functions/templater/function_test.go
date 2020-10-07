package templater

import (
	"testing"

	"bytes"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestTemplater(t *testing.T) {

	tc := []struct {
		cfg         string
		in          string
		expectedOut string
		expectedErr bool
	}{
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  hosts:
    - macAddress: 00:aa:bb:cc:dd
      name: node-1
    - macAddress: 00:aa:bb:cc:ee
      name: node-2
template: |
  {{ range .hosts -}}
  ---
  apiVersion: metal3.io/v1alpha1
  kind: BareMetalHost
  metadata:
    name: {{ .name }}
  spec:
    bootMACAddress: {{ .macAddress }}
  {{ end -}}
`,
			expectedOut: `apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-1
spec:
  bootMACAddress: 00:aa:bb:cc:dd
---
apiVersion: metal3.io/v1alpha1
kind: BareMetalHost
metadata:
  name: node-2
spec:
  bootMACAddress: 00:aa:bb:cc:ee
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  test:
    of:
      - toYaml
template: |
  {{ toYaml . -}}
`,
			expectedOut: `test:
  of:
  - toYaml
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
values:
  test:
    of:
      - badToYamlInput
template: |
  {{ toYaml ignorethisbadinput -}}
`,
			expectedErr: true,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
template: |
  {{ end }`,
			expectedErr: true,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
template: |
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: map2
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
`,
		},
		{
			in: `
apiVersion: v1
kind: ConfigMap
metadata:
  name: map1
`,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: Templater
metadata:
  name: notImportantHere
cleanPipeline: true
template: |
  apiVersion: v1
  kind: ConfigMap
  metadata:
    name: map2
`,
			expectedOut: `apiVersion: v1
kind: ConfigMap
metadata:
  name: map2
`,
		},
	}

	for i, ti := range tc {
		fcfg := FunctionConfig{}
		err := yaml.Unmarshal([]byte(ti.cfg), &fcfg)
		if err != nil {
			t.Errorf("can't unmarshal config %s. continue", ti.cfg)
			continue
		}

		nodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(ti.in)}).Read()
		if err != nil {
			t.Errorf("can't unmarshal config %s. continue", ti.in)
			continue
		}

		f := Function{Config: &fcfg}
		nodes, err = f.Exec(nodes)
		if err != nil && !ti.expectedErr {
			t.Errorf("exec %d returned unexpected error %v for %s", i, err, ti.cfg)
			continue
		}
		out := &bytes.Buffer{}
		err = kio.ByteWriter{Writer: out}.Write(nodes)
		if err != nil {
			t.Errorf("write returned unexpected error %v for %s", err, ti.cfg)
			continue
		}
		if out.String() != ti.expectedOut {
			t.Errorf("expected %s, got %s", ti.expectedOut, out.String())
		}
	}

}
