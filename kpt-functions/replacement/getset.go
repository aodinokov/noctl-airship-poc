package replacement

import (
	//"log"

	"bytes"
	"fmt"
	"strconv"
	"unicode/utf8"

	"sigs.k8s.io/kustomize/kyaml/yaml"
)

func ParseFieldRefs(in string) ([]string, error) {
	var cur bytes.Buffer
	out := []string{}
	var state int
	for i := 0; i < len(in); {
		r, size := utf8.DecodeRuneInString(in[i:])

		switch state {
		case 0: // initial state
			if r == '|' {
				if cur.String() != "" {
					out = append(out, cur.String())
					cur = bytes.Buffer{}
				}
			} else if r == '[' {
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

func ParseFieldRef(in string) ([]string, error) {
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

func seqNodeIndexPath(p string) (int64, error) {
	if p[0] == '[' && p[len(p)-1] == ']' {
		p = p[1 : len(p)-1]
	}
	i, err := strconv.ParseInt(p, 10, 64)
	if err != nil {
		return 0, err
	}
	if i < 0 {
		return 0, fmt.Errorf("index can't be negative. got %d", i)
	}
	return i, nil
}

func getFieldValue(node *yaml.RNode, fieldRef string) (interface{}, error) {
	fieldRefs, err := ParseFieldRefs(fieldRef)
	if err != nil {
		return nil, err
	}

	node, err = getFieldValueImpl(node, fieldRefs)
	if err != nil {
		return nil, err
	}
	if node.YNode().Kind == yaml.ScalarNode {
		return yaml.GetValue(node), nil
	}

	return node, nil
}

func getFieldValueImpl(node *yaml.RNode, fieldRefs []string) (*yaml.RNode, error) {
	path, err := ParseFieldRef(fieldRefs[0])
	if err != nil {
		return nil, err
	}

	cn := node
	for _, p := range path {

		// index case
		if cn.YNode().Kind == yaml.SequenceNode {
			i, err := seqNodeIndexPath(p)

			//log.Printf("get: found index %d", i)

			if err == nil {
				content := cn.Content()
				if i >= int64(len(content)) {
					return nil, fmt.Errorf("index %d is too big", i)
				}
				cn = yaml.NewRNode(content[i])
				continue
			}
		}

		// default case - use lookup
		cn, err = cn.Pipe(yaml.Lookup(p))
		if err != nil {
			return nil, err
		}
	}

	if len(fieldRefs) == 1 {
		return cn, nil
	}

	if cn.YNode().Kind != yaml.ScalarNode {
		return nil, fmt.Errorf("node %v isn't scalar", path)
	}

	node, err = yaml.Parse(yaml.GetValue(cn))
	if err != nil {
		return nil, err
	}

	return getFieldValueImpl(node, fieldRefs[1:])
}

func guessNodeKind(i int, path []string, node *yaml.RNode) (yaml.Kind, error) {
	if i >= len(path) || i < 0 {
		return yaml.SequenceNode, fmt.Errorf("incorrect i %d", i)
	}

	if i == len(path)-1 {
		return node.YNode().Kind, nil
	}

	p := path[i+1]
	_, err := seqNodeIndexPath(p)
	if err == nil || (p[0] == '[' && p[len(p)-1] == ']') {
		return yaml.SequenceNode, nil
	}

	return yaml.MappingNode, nil
}

func setFieldValue(node *yaml.RNode, fieldRef string, value interface{}) error {
	var setNode *yaml.RNode

	svalue, ok := value.(string)
	if ok {
		setNode = yaml.NewScalarRNode(svalue)
	} else {
		setNode, ok = value.(*yaml.RNode)
		if !ok {
			return fmt.Errorf("value arg containes not expected type")
		}
	}

	fieldRefs, err := ParseFieldRefs(fieldRef)
	if err != nil {
		return err
	}

	return setFieldValueImpl(node, fieldRefs, setNode)
}

func setFieldValueImpl(node *yaml.RNode, fieldRefs []string, setNode *yaml.RNode) error {
	path, err := ParseFieldRef(fieldRefs[0])
	if err != nil {
		return err
	}

	cn := node
	for i, p := range path {
		// index case
		if cn.YNode().Kind == yaml.SequenceNode {
			indx, err := seqNodeIndexPath(p)
			//log.Printf("set: found index %d", indx)
			if err == nil {
				content := cn.Content()
				if indx >= int64(len(content)) {
					// we don't create by index
					return fmt.Errorf("index %d is too big", indx)
				}
				//log.Printf("building content: %v", content[indx])
				if i < len(path)-1 {
					cn = yaml.NewRNode(content[indx])
					continue
				}
				child := yaml.NewRNode(content[indx])
				if child.YNode().Kind == yaml.ScalarNode {
					// converting index to brackets format
					p = fmt.Sprintf("[=%s]", yaml.GetValue(child))
				} else {
					return fmt.Errorf("don't support case when it's necesary to find uniq param")
				}
			}
		}

		kind, err := guessNodeKind(i, path, setNode)
		if err != nil {
			return fmt.Errorf("wasn't able to guess node kind: %v", err)
		}
		// override to saclar if there is included yaml
		if i == len(path)-1 && len(fieldRefs) > 1 {
			kind = yaml.ScalarNode
		}

		// default case - use lookup
		cnl, err := cn.Pipe(yaml.Lookup(p))
		if err != nil {
			return fmt.Errorf("wan't able to lookup %v", err)
		}
		if cnl == nil {
			cnl, err = cn.Pipe(yaml.LookupCreate(kind, p))
			if err != nil {
				return fmt.Errorf("wan't able to create node %v", err)
			}
			if cnl == nil {
				return fmt.Errorf("unexpected nil object pointer returned")
			}
		} else {
			if cnl.YNode().Kind != kind {
				if cnl.YNode().Kind == yaml.ScalarNode && yaml.GetValue(cnl) == "" {
					//TODO: change
					return fmt.Errorf("unexpected kind in %v. possible change from emptyScalar isn't implemented", path[:i+1])
				} else {
					return fmt.Errorf("unexpected kind in %v", path[:i+1])
				}
			}
		}

		cnp := cn
		cn = cnl

		if i == len(path)-1 {
			if len(fieldRefs) > 1 {
				includedNode, err := yaml.Parse(yaml.GetValue(cn))
				if err != nil {
					return fmt.Errorf("wan't able to parse %s", yaml.GetValue(cn))
				}
				err = setFieldValueImpl(includedNode, fieldRefs[1:], setNode)
				if err != nil {
					return fmt.Errorf("wan't able to setFieldValueImpl %v", err)
				}
				s, err := includedNode.String()
				if err != nil {
					return fmt.Errorf("wan't able to convert to string %v", err)
				}
				err = cn.PipeE(yaml.FieldSetter{StringValue: s})
				if err != nil {
					return fmt.Errorf("wan't able to set back: %v", err)
				}
			} else {
				if cnp.YNode().Kind == yaml.MappingNode {
					err = cnp.PipeE(yaml.FieldSetter{Name: p, Value: setNode})
					if err != nil {
						return fmt.Errorf("wan't able to set map: %v", err)
					}
				} else { /*opposite is only yaml.SequenceNode */
					// we need to delete the found element
					// and set the new one instead
					k, v, err := yaml.SplitIndexNameValue(p)
					if err != nil {
						return fmt.Errorf("can't get kv %s", p)
					}

					err = cnp.PipeE(yaml.ElementSetter{Element: setNode.YNode(), Key: k, Value: v})
					if err != nil {
						return fmt.Errorf("wan't able to set seq: %v", err)
					}
				}
			}
		}
	}

	return nil
}
