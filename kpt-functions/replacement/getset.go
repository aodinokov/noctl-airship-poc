package replacement

import (
	"fmt"
	"strings"

	"log"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func getFieldValue(node *yaml.RNode, fieldRef string) (interface{}, error) {
	var value interface{}
	parts := strings.Split(fieldRef, "|")
	for i, path := range parts {
		v, err := node.Pipe(yaml.PathGetter{Path: strings.Split(path, ".")})
		if err != nil {
			return nil, err
		}
		if v != nil {
			value = v

			if v.YNode().Kind == yaml.ScalarNode {
				value = yaml.GetValue(v)
			}
			if i+1 < len(parts) {
				if v.YNode().Kind != yaml.ScalarNode {
					return nil, fmt.Errorf("node %v isn't scalar", path)
				}
				node, err = yaml.Parse(value.(string))
				if err != nil {
					return nil, err
				}
			}
		}
	}
	return value, nil
}

func setFieldValue(node *yaml.RNode, fieldRef string, value interface{}) error {
	return setFieldValueImpl(node, strings.Split(fieldRef, "|"), value)
}

func setFieldValueImpl(node *yaml.RNode, fieldRefPart []string, value interface{}) error {
	//ds, _ := node.String()
	//log.Printf("setFieldValueImpl %s %v %s", ds, fieldRefPart, value)
	//defer log.Printf("exit setFieldValueImpl %s %v %s", ds, fieldRefPart, value)
	if len(fieldRefPart) > 1 {
		// this can be done only for string field
		//v, err := node.Pipe(yaml.Lookup(strings.Split(fieldRefPart[0], ".")...))
		v, err := node.Pipe(yaml.PathGetter{Path: strings.Split(fieldRefPart[0], ".")})
		if err != nil {
			return fmt.Errorf("wasn't able to lookup %s: %w", fieldRefPart[0], err)
		}
		if v == nil {
			return fmt.Errorf("wasn't able to find value for fieldref %s", fieldRefPart[0])
		}
		//log.Printf("parsing %s", yaml.GetValue(v))
		includedNode, err := yaml.Parse(yaml.GetValue(v))
		if err != nil {
			return fmt.Errorf("wasn't able to parse yaml value for fieldref %s", fieldRefPart[0])
		}
		err = setFieldValueImpl(includedNode, fieldRefPart[1:], value)
		if err != nil {
			return fmt.Errorf("recursive %s: %w", fieldRefPart[0], err)
		}
		s, err := includedNode.String()
		if err != nil {
			return fmt.Errorf("can't marshal includedNode: %w", err)
		}
		//log.Printf("setting %s", s)
		err = v.PipeE(yaml.FieldSetter{StringValue: s})
		if err != nil {
			return fmt.Errorf("can't set new value %s back: %w", s, err)
		}
		return nil

	}

	svalue, ok := value.(string)
	if ok {
		log.Printf("looking for %s", fieldRefPart[0])
		v, err := node.Pipe(yaml.LookupCreate(yaml.ScalarNode, strings.Split(fieldRefPart[0], ".")...))
		if err != nil {
			return fmt.Errorf("scalar case: wasn't able to lookup %v: %w", strings.Split(fieldRefPart[0], "."), err)
		}
		log.Printf("found %s", yaml.GetValue(v))
		err = v.PipeE(yaml.FieldSetter{StringValue: svalue})
		if err != nil {
			return fmt.Errorf("scalar case: fieldsetter returned error for %s: %w", fieldRefPart[0], err)
		}
		return nil
	}
	rnode, ok := value.(*yaml.RNode)
	if ok {
		path := strings.Split(fieldRefPart[0], ".")
		v := node
		if len(path) > 1 {
			var err error
			v, err = node.Pipe(yaml.Lookup(path[:len(path)-1]...))
			if err != nil {
				return fmt.Errorf("wasn't able to lookup %s: %w", fieldRefPart[0], err)
			}
		}
		log.Printf("setting Name %v to %v", path[len(path)-1], path[:len(path)-1])
		err := v.PipeE(yaml.FieldSetter{Name: path[len(path)-1], Value: rnode})
		if err != nil {
			return fmt.Errorf("fieldsetter returned error for %s: %w", fieldRefPart[0], err)
		}
		return nil
	}
	return fmt.Errorf("unexpected value type %v: %T", value, value)
}
