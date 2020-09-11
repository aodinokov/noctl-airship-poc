package replacement

import (
	"testing"

	"bytes"

	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func TestReplacements(t *testing.T) {
	tc := []struct {
		cfg         string
		in          string
		expectedOut string
		expectedErr bool
	}{
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: nginx:newtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[name=nginxlatest].image
- source:
    value: postgres:latest
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[name=postgresdb].image
`,

			in: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:latest
        name: nginxlatest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
			expectedOut: `group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:newtag
        name: nginxlatest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:latest
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: Pod
      name: pod
    fieldref: spec.containers
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers`,
			in: `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
`,
			expectedOut: `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - name: myapp-container
        image: busybox
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy3
spec:
  template:
    spec:
      containers:
      - name: myapp-container
        image: busybox
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: ConfigMap
      name: cm
    fieldref: data.HOSTNAME
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[image=debian].args.[=HOSTNAME]
    - spec.template.spec.containers.[name=busybox].args.[=HOSTNAME]
- source:
    objref:
      kind: ConfigMap
      name: cm
    fieldref: data.PORT
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[image=debian].args.[=PORT]
    - spec.template.spec.containers.[name=busybox].args.[=PORT]`,
			in: `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  labels:
    foo: bar
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
        - name: command-demo-container
          image: debian
          command: ["printenv"]
          args:
            - HOSTNAME
            - PORT
        - name: busybox
          image: busybox:latest
          args:
            - echo
            - HOSTNAME
            - PORT
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOSTNAME: example.com
  PORT: 8080`,
			expectedOut: `apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy
  labels:
    foo: bar
spec:
  template:
    metadata:
      labels:
        foo: bar
    spec:
      containers:
      - name: command-demo-container
        image: debian
        command: ["printenv"]
        args:
        - example.com
        - 8080
      - name: busybox
        image: busybox:latest
        args:
        - echo
        - example.com
        - 8080
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: cm
data:
  HOSTNAME: example.com
  PORT: 8080
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    value: regexedtag
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[name=nginx-latest].image%TAG%
- source:
    value: postgres:latest
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[name=postgresdb].image`,
			in: `
group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:TAG
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:1.8.0
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine`,
			expectedOut: `group: apps
apiVersion: v1
kind: Deployment
metadata:
  name: deploy1
spec:
  template:
    spec:
      containers:
      - image: nginx:1.7.9
        name: nginx-tagged
      - image: nginx:regexedtag
        name: nginx-latest
      - image: foobar:1
        name: replaced-with-digest
      - image: postgres:latest
        name: postgresdb
      initContainers:
      - image: nginx
        name: nginx-notag
      - image: nginx@sha256:111111111111111111
        name: nginx-sha256
      - image: alpine:1.8.0
        name: init-alpine
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: Pod
      name: pod1
  target:
    objref:
      kind: Pod
      name: pod2
    fieldrefs:
    - spec.non.existent.field`,
			in: `
apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  containers:
  - name: myapp-container
    image: busybox`,
			expectedOut: `apiVersion: v1
kind: Pod
metadata:
  name: pod1
spec:
  containers:
  - name: myapp-container
    image: busybox
---
apiVersion: v1
kind: Pod
metadata:
  name: pod2
spec:
  containers:
  - name: myapp-container
    image: busybox
  non:
    existent:
      field: pod1
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: Pod
      name: pod
    fieldref: spec.containers.[name=repl]
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers.[name=myapp-container]`,
			in: `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: repl
    image: repl
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox
        name: myapp-container
`,
			expectedOut: `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: repl
    image: repl
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - name: repl
        image: repl
`,
		},
		// Additional feature: yaml in yaml
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: VariableCatalogue
      name: source
    fieldref: values.field1
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.userData|bootcmd.[=mkdir /mnt/vda || echo error]`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda || echo error
`,
			expectedOut: `apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - value1
`,
		},
		// Additional feature: multiref stringbuilder
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    multiref:
      refs:
      - objref:
          kind: VariableCatalogue
          name: source
        fieldref: values.image
      - objref:
          kind: VariableCatalogue
          name: source
        fieldref: values.tag
      template: |-
        {{ index .Values 0 }}:{{ index .Values 1 }}
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.image`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  image: imagevalue
  tag: tagvalue
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  image: someimage:sometag
`,
			expectedOut: `apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  image: imagevalue
  tag: tagvalue
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  image: imagevalue:tagvalue
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    multiref:
      refs:
      - objref:
          kind: Secret
          name: node1-bmc-secret
        fieldref: stringData.image
      - objref:
          kind: VariableCatalogue
          name: source
        fieldref: values.tag
      template: |-
        {{ regexReplaceAll ":.*$" (index .Values 0) (printf ":%s" (index .Values 1)) }}
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.image`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  image: imagevalue
  tag: tagvalue
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  image: someimage:sometag
`,
			expectedOut: `apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  image: imagevalue
  tag: tagvalue
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  image: someimage:tagvalue
`,
		},
		// digit indexing must also work
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: VariableCatalogue
      name: source
    fieldref: values.field1
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.userData|bootcmd.[0]`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - mkdir /mnt/vda
`,
			expectedOut: `apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    bootcmd:
    - value1
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: VariableCatalogue
      name: source
    fieldref: values.field1
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.userData|write_files.[path=/etc/kubernetes/admin.conf].content|apiVersion`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    write_files:
    - content: |
        apiVersion: v1
      path: /etc/kubernetes/admin.conf
`,
			expectedOut: `apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    write_files:
    - content: |
        apiVersion: value1
      path: /etc/kubernetes/admin.conf
`,
		},
		// Negative test for target incorrct fieldref (must be an eeor)
		{
			expectedErr: true,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: VariableCatalogue
      name: source
    fieldref: values.field1
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.userData|write_files.[path=/etc/kubernetes/nofile.conf].content|apiVersion`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    write_files:
    - content: |
        apiVersion: v1
      path: /etc/kubernetes/admin.conf
`,
		},
		// negative for incorrect source fieldref
		{
			expectedErr: true,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    objref:
      kind: VariableCatalogue
      name: source
    fieldref: values.field3
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.userData|write_files.[path=/etc/kubernetes/admin.conf].content|apiVersion`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    write_files:
    - content: |
        apiVersion: v1
      path: /etc/kubernetes/admin.conf
`,
		},
		// negative for incorrect source fieldref for multiref
		{
			expectedErr: true,
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: notImportantHere
replacements:
- source:
    multiref:
      refs:
      - objref:
          kind: VariableCatalogue
          name: source
        fieldref: values.field3
    template: |-
      {{ index .Values 0 }}-posfix
  target:
    objref:
      kind: Secret
    fieldrefs:
    - stringData.userData|write_files.[path=/etc/kubernetes/admin.conf].content|apiVersion`,
			in: `
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
  name: source
values:
  field1: value1
  field2: value2
---
apiVersion: v1
kind: Secret
metadata:
  name: node1-bmc-secret
type: Opaque
stringData:
  userData: |
    write_files:
    - content: |
        apiVersion: v1
      path: /etc/kubernetes/admin.conf
`,
		},
		{
			cfg: `
apiVersion: airshipit.org/v1alpha1
kind: ReplacementTransformer
metadata:
  name: test-for-numeric-conversion
replacements:
- source:
    objref:
      kind: Pod
      name: pod
    fieldref: spec.containers[0].image
  target:
    objref:
      kind: Deployment
    fieldrefs:
    - spec.template.spec.containers[name=myapp-container].image%TAG%`,
			in: `
apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: repl
    image: 12345
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox:TAG
        name: myapp-container
`,
			expectedOut: `apiVersion: v1
kind: Pod
metadata:
  name: pod
spec:
  containers:
  - name: repl
    image: 12345
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: deploy2
spec:
  template:
    spec:
      containers:
      - image: busybox:12345
        name: myapp-container
`,
		},
	}

	for i, ti := range tc {
		fcfg := FunctionConfig{}
		err := yaml.Unmarshal([]byte(ti.cfg), &fcfg)
		if err != nil {
			t.Errorf("can't unmarshal config %s: %v continue", ti.cfg, err)
			continue
		}

		nodes, err := (&kio.ByteReader{Reader: bytes.NewBufferString(ti.in)}).Read()
		if err != nil {
			t.Errorf("can't unmarshal config %s. continue", ti.in)
			continue
		}

		f := Function{Config: &fcfg}
		err = f.Exec(nodes)
		if err != nil && !ti.expectedErr {
			t.Errorf("exec %d returned unexpected error %v for %s", i, err, ti.cfg)
			continue
		}
		if ti.expectedErr {
			//t.Logf("expected error, msg %v", err)
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
