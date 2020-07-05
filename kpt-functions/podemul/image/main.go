// Package main implements pod emulation function to run arbitrary scripts and
// is run with `kustomize config run -- DIR/`.
package main

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/google/shlex"

	"sigs.k8s.io/kustomize/kyaml/fn/framework"
	"sigs.k8s.io/kustomize/kyaml/yaml"
)

// Copy of Pod structure elements, but
// instead of container we're using executabe (because we already in container)
type Executable struct {
	// name of executable
	Name string `yaml:"name"`
	// cmdline to execute
	Cmdline string `yaml:"cmdline"`
	// set of volumes and their mount points
	VolumeMounts []struct {
		// reference to volume
		Name string `yaml:"name"`
		// its mount point
		MountPath string `yaml:"mountPath"`
	} `yaml:"volumeMounts,omitempty"`
	// env variables to be configured
	Env []struct {
		// env var name
		Name string `yaml:"name"`
		// its value
		Value *string `yaml:"value,omitempty"`
		// ... or its value taken from
		ValueFrom *struct {
			// .... configmap
			ConfigMapKeyRef *struct {
				// confirmpa name
				Name string `yaml:"name"`
				// key inside of configmap
				Key string `yaml:"key"`
			} `yaml:"configMapKeyRef,omitempty"`
			// ... secret
			SecretKeyRef *struct {
				// secret name
				Name string `yaml:"name"`
				// key inside of secret
				Key string `yaml:"key"`
			} `yaml:"secretKeyRef,omitempty"`
		} `yaml:"valueFrom,omitempty"`
	} `yaml:"env,omitempty"`
}

// Volume and its subtypes
type VolumeItem struct {
	// item key
	Key string `yaml:"key"`
	// item permission override
	Mode *os.FileMode `yaml:"mode,omitempty"`
	// item path/filename override
	Path *string `yaml:"path,omitempty"`
}

// data scruct to keep Volume configuration
type Volume struct {
	// Volume name
	Name string `yaml:"name"`
	// configMap reference (may be ommitted)
	ConfigMap *struct {
		// name of referenced configmap
		Name string `yaml:"name"`
		// default file permission
		DefaultMode *os.FileMode `yaml:"defaultMode,omitempty"`
		// per-key configuration
		Items []VolumeItem `yaml:"items,omitempty"`
	} `yaml:"configMap,omitempty"`
	// secret reference (may be ommitted)
	Secret *struct {
		// name of referenced secret
		Name string `yaml:"name"`
		// default file permission
		DefaultMode *os.FileMode `yaml:"defaultMode,omitempty"`
		// per-key configuration
		Items []VolumeItem `yaml:"items,omitempty"`
	} `yaml:"secret,omitempty"`
	// map to get item by its key
	ItemsMap map[string]*VolumeItem
}

// Key value is used to emulate
// secrets and config map
type KeyValue struct {
	// Data keeps all configMap/Secret keys and
	// corresponsing values
	Data map[string][]byte
}

// define the input API schema as a struct
type Function struct {
	// Function metadata
	Metadata struct {
		// Function namespace
		// used to filter ConfigMaps and Secrets
		Namespace string `yaml:"namespace,omitempty"`
	} `yaml:"metadata,omitempty"`
	// function spec
	Spec struct {
		// array of executalbes configuration
		Executables []Executable `yaml:"executables,omitempty"`
		// arrays of volumes - shared between executables
		// each file is recreated for each executable
		Volumes []Volume `yaml:"volumes,omitempty"`
	} `yaml:"spec"`

	// volume to volume.name map
	VolumesMap map[string]*Volume
	// list of all configmaps from the same namespace as function
	ConfigMaps map[string]*KeyValue
	// list of all secrets from the same namespace as function
	Secrets map[string]*KeyValue
}

func main() {
	log.Print("started")
	defer log.Print("Finished")

	function := &Function{
		VolumesMap: map[string]*Volume{},
		ConfigMaps: map[string]*KeyValue{},
		Secrets:    map[string]*KeyValue{}}
	resourceList := &framework.ResourceList{FunctionConfig: function}

	cmd := framework.Command(resourceList, func() error {
		err := function.FinalizeInit()
		if err != nil {
			log.Printf("function initialization failed: %v", err)
			return err
		}

		for _, r := range resourceList.Items {
			if err := function.Scan(r); err != nil {
				log.Printf("error %v", err)
				return err
			}
		}

		return function.Exec()
	})

	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

// checks the structure invariant after unmarshaling
// assignes some default values if some data is missed
func (f *Function) FinalizeInit() error {
	if f.Metadata.Namespace == "" {
		f.Metadata.Namespace = "default"
	}

	for i := range f.Spec.Volumes {
		if err := (&f.Spec.Volumes[i]).FinalizeInit(); err != nil {
			return err
		}
		// Make some caching
		f.VolumesMap[f.Spec.Volumes[i].Name] = &f.Spec.Volumes[i]
	}

	for i := range f.Spec.Executables {
		if err := (&f.Spec.Executables[i]).FinalizeInit(); err != nil {
			return err
		}
	}
	return nil
}

func (v *Volume) FinalizeInit() error {
	v.ItemsMap = map[string]*VolumeItem{}

	if v.ConfigMap == nil && v.Secret == nil {
		return fmt.Errorf("volume %s has to specify configMap or secret", v.Name)
	}
	if v.ConfigMap != nil && v.Secret != nil {
		return fmt.Errorf("volume %s has to specify either configMap or secret, but not both", v.Name)
	}
	// Make some chaching
	if v.ConfigMap != nil {
		for i := range v.ConfigMap.Items {
			v.ItemsMap[v.ConfigMap.Items[i].Key] = &v.ConfigMap.Items[i]
		}
	}
	if v.Secret != nil {
		for i := range v.Secret.Items {
			v.ItemsMap[v.Secret.Items[i].Key] = &v.Secret.Items[i]
		}
	}
	return nil
}

func (e *Executable) FinalizeInit() error {
	for _, v := range e.Env {
		if v.Value == nil && v.ValueFrom == nil {
			return fmt.Errorf("env %s has to specify value or valueFrom", v.Name)
		}
		if v.Value != nil && v.ValueFrom != nil {
			return fmt.Errorf("env %s has to specify value or valueFrom, but not both", v.Name)
		}

		if v.ValueFrom != nil {
			if v.ValueFrom.ConfigMapKeyRef == nil && v.ValueFrom.SecretKeyRef == nil {
				return fmt.Errorf("env %s has to specify configMapKeyRef or secretKeyRef in valueFrom", v.Name)
			}
			if v.ValueFrom.ConfigMapKeyRef != nil && v.ValueFrom.SecretKeyRef != nil {
				return fmt.Errorf("env %s has to specify configMapKeyRef or secretKeyRef in valueFrom, but not both", v.Name)
			}
		}
	}
	return nil
}

// KV constructor from ConfigMap
func NewKeyValueFromConfigMap(r *yaml.RNode) (*KeyValue, error) {
	data, err := r.Pipe(yaml.Lookup("data"))
	if err != nil {
		s, _ := r.String()
		return nil, fmt.Errorf("%v: %s", err, s)
	}

	if data == nil {
		s, _ := r.String()
		return nil, fmt.Errorf("no data field: %s", s)
	}

	fields, err := data.Fields()
	if err != nil {
		s, _ := r.String()
		return nil, fmt.Errorf("couldn't take data fields: %v, %s", err, s)
	}

	kv := KeyValue{Data: map[string][]byte{}}
	for _, field := range fields {
		valueRNode := data.Field(field).Value
		kv.Data[field] = []byte(yaml.GetValue(valueRNode))
	}

	return &kv, nil
}

// KV constructor from Secret
func NewKeyValueFromSecret(r *yaml.RNode) (*KeyValue, error) {
	kv := KeyValue{Data: map[string][]byte{}}
	// Decode base64
	data, err := r.Pipe(yaml.Lookup("data"))
	if err != nil {
		s, _ := r.String()
		return nil, fmt.Errorf("%v: %s", err, s)
	}
	if data != nil {
		fields, err := data.Fields()
		if err != nil {
			s, _ := r.String()
			return nil, fmt.Errorf("couldn't take data fields: %v, %s", err, s)
		}
		for _, field := range fields {
			valueB64 := yaml.GetValue(data.Field(field).Value)
			decodedData, err := base64.StdEncoding.DecodeString(valueB64)
			if err != nil {
				return nil, err
			}
			kv.Data[field] = decodedData
		}
	}
	// Decode stringData
	data, err = r.Pipe(yaml.Lookup("stringData"))
	if err != nil {
		s, _ := r.String()
		return nil, fmt.Errorf("%v: %s", err, s)
	}
	if data != nil {
		fields, err := data.Fields()
		if err != nil {
			s, _ := r.String()
			return nil, fmt.Errorf("couldn't take stringData fields: %v, %s", err, s)
		}
		for _, field := range fields {
			if _, ok := kv.Data[field]; ok {
				return nil, fmt.Errorf("duplicated field %s present in data and in stringData", field)
			}
			valueRNode := data.Field(field).Value
			kv.Data[field] = []byte(yaml.GetValue(valueRNode))
		}
	}
	return &kv, nil
}

// what is done:
// done 1. make the ame for Secret for both data and stringData
// done 2. uodate Scan function to add all matching configMaps and secrets to f
// done 3. make fn that builds env based on what executable has in config
// done 4. make function that creates needed files based on volume config and returns the object
// done 5. this object has only 1 function - to remove all files after exec
// done 6  update main - after scan we have to walk though each of exec and do: volumes, env, cmdline, removeFiles,

// Scan ResourceList element for userData and networkData
func (f *Function) Scan(r *yaml.RNode) error {
	meta, err := r.GetMeta()
	if err != nil {
		return err
	}

	if meta.Kind != "ConfigMap" && meta.Kind != "Secret" {
		return nil
	}

	namespace := "default"
	if meta.Namespace != "" {
		namespace = meta.Namespace
	}

	if namespace != f.Metadata.Namespace {
		// skipping resource because of different namespace
		return nil
	}

	switch meta.Kind {
	case "ConfigMap":
		if _, ok := f.ConfigMaps[meta.Name]; ok {
			return fmt.Errorf("error: trying to add the second ConfigMap with the same name %s", meta.Name)
		}
		cm, err := NewKeyValueFromConfigMap(r)
		if err != nil {
			return err
		}
		f.ConfigMaps[meta.Name] = cm
	case "Secret":
		if _, ok := f.Secrets[meta.Name]; ok {
			return fmt.Errorf("error: trying to add the second Secret with the same name %s", meta.Name)
		}
		s, err := NewKeyValueFromSecret(r)
		if err != nil {
			return err
		}
		f.Secrets[meta.Name] = s
	}

	return nil
}

func (f *Function) Exec() error {
	for i := range f.Spec.Executables {
		if err := f.exec(&f.Spec.Executables[i]); err != nil {
			return err
		}
	}
	return nil
}

func ensureDir(path string, createMode os.FileMode) ([]string, error) {
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		return []string{}, nil
	}

	createdVolumes, err := ensureDir(filepath.Dir(path), createMode)
	if err != nil {
		return createdVolumes, err
	}

	err = os.Mkdir(path, createMode)
	if err != nil {
		return createdVolumes, err
	}

	return append(createdVolumes, path), nil
}

func (f *Function) initVolume(
	kv *KeyValue,
	mountPath string,
	defaultMode *os.FileMode,
	itemsMap map[string]*VolumeItem) ([]string, error) {

	// Calculate the default mode
	dm := os.FileMode(0640)
	if defaultMode != nil {
		dm = *defaultMode
	}

	createdVolumes := []string{}
	for key, val := range kv.Data {
		// Calcuate the mode and path
		mode := dm
		path := key
		// Get overrides
		item, ok := itemsMap[key]
		if ok {
			if item.Mode != nil {
				mode = *item.Mode
			}
			if item.Path != nil {
				path = *item.Path
			}
		}
		// update path with mount path
		path = filepath.Join(mountPath, path)

		// make sure dir is created
		// if not - create and add it to createdVolumes
		// TODO: 0750 is real defautl? should we set this in config?
		cv, err := ensureDir(filepath.Dir(path), 0750)
		createdVolumes = append(createdVolumes, cv...)
		if err != nil {
			return createdVolumes, err
		}

		// create file
		err = ioutil.WriteFile(path, val, mode)
		if err != nil {
			return createdVolumes, err
		}
		// keep data that we created file
		createdVolumes = append(createdVolumes, path)
	}

	return createdVolumes, nil
}

func (f *Function) initVolumes(e *Executable) ([]string, error) {
	createdVolumes := []string{}
	for _, vm := range e.VolumeMounts {
		volume, ok := f.VolumesMap[vm.Name]
		if !ok {
			return createdVolumes, fmt.Errorf("volume %s wasn't found in %v", vm.Name, f.VolumesMap)
		}

		if volume.ConfigMap != nil {
			kv, ok := f.ConfigMaps[volume.ConfigMap.Name]
			if !ok {
				return createdVolumes, fmt.Errorf("ConfigMap %s used in volume %s wasn't found",
					volume.ConfigMap.Name, vm.Name)
			}
			cv, err := f.initVolume(kv, vm.MountPath, volume.ConfigMap.DefaultMode, volume.ItemsMap)
			createdVolumes = append(createdVolumes, cv...)
			if err != nil {
				return createdVolumes, err
			}
		} else if volume.Secret != nil {
			kv, ok := f.Secrets[volume.Secret.Name]
			if !ok {
				return createdVolumes, fmt.Errorf("Secret %s used in volume %s wasn't found",
					volume.Secret.Name, vm.Name)
			}
			cv, err := f.initVolume(kv, vm.MountPath, volume.Secret.DefaultMode, volume.ItemsMap)
			createdVolumes = append(createdVolumes, cv...)
			if err != nil {
				return createdVolumes, err
			}
		} else {
			return createdVolumes, fmt.Errorf("volume %s has unitialized ConfigMap and Secret")
		}
	}
	return createdVolumes, nil
}

func (f *Function) uninitVolumes(createdVolumes []string) error {
	nonDeletedVolumes := []string{}
	for i := range createdVolumes {
		// process it in reverse order,
		// so dirs will be deleted after files
		path := createdVolumes[len(createdVolumes)-i-1]
		err := os.Remove(path)
		if err != nil {
			nonDeletedVolumes = append(nonDeletedVolumes, path)
		}
	}
	if len(nonDeletedVolumes) > 0 {
		return fmt.Errorf("some files or dirs were not removed: %v", nonDeletedVolumes)
	}
	return nil
}

func (f *Function) getEnv(e *Executable) ([]string, error) {
	env := []string{}
	for i := range e.Env {
		if e.Env[i].Value != nil {
			env = append(env, fmt.Sprintf("%s=%s", e.Env[i].Name, *e.Env[i].Value))
			continue
		}
		if e.Env[i].ValueFrom.ConfigMapKeyRef != nil {
			kv, ok := f.ConfigMaps[e.Env[i].ValueFrom.ConfigMapKeyRef.Name]
			if !ok {
				return nil, fmt.Errorf("can't find ConfigMap with name %s", e.Env[i].ValueFrom.ConfigMapKeyRef.Name)
			}
			val, ok := kv.Data[e.Env[i].ValueFrom.ConfigMapKeyRef.Key]
			if !ok {
				return nil, fmt.Errorf("can't find key with name %s in ConfigMap %s",
					e.Env[i].ValueFrom.ConfigMapKeyRef.Key,
					e.Env[i].ValueFrom.ConfigMapKeyRef.Name)
			}
			env = append(env, fmt.Sprintf("%s=%s", e.Env[i].Name, string(val)))
			continue
		}
		if e.Env[i].ValueFrom.SecretKeyRef != nil {
			kv, ok := f.Secrets[e.Env[i].ValueFrom.SecretKeyRef.Name]
			if !ok {
				return nil, fmt.Errorf("can't find Secret with name %s", e.Env[i].ValueFrom.SecretKeyRef.Name)
			}
			val, ok := kv.Data[e.Env[i].ValueFrom.SecretKeyRef.Key]
			if !ok {
				return nil, fmt.Errorf("can't find key with name %s in Secret %s",
					e.Env[i].ValueFrom.SecretKeyRef.Name,
					e.Env[i].ValueFrom.SecretKeyRef.Key)
			}
			env = append(env, fmt.Sprintf("%s=%s", e.Env[i].Name, string(val)))
			continue
		}
		return nil, fmt.Errorf("env item %s(%d) isn't initialize properly", e.Env[i].Name, i)
	}
	return env, nil
}

func (f *Function) exec(e *Executable) error {
	env, err := f.getEnv(e)
	if err != nil {
		return err
	}

	createdVolumes, err := f.initVolumes(e)
	defer func() {
		err := f.uninitVolumes(createdVolumes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "uninitVolumes returned error: %v", err)
		}
	}()
	if err != nil {
		return err
	}

	args, err := shlex.Split(e.Cmdline)
	if err != nil {
		return err
	}
	if len(args) < 1 {
		return fmt.Errorf("incorrect cmdline %s", e.Cmdline)
	}

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(), env...)

	return cmd.Run()
}
