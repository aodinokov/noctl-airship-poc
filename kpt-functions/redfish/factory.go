package redfish

import(
	"fmt"
	"regexp"
)

type Driver interface {
	// TODO:
}

type Model struct {
	// if empty - when we don't check - good for default values
	Re *regexp.Regexp

	// constructor
	Constructor func() (*Driver, error)
}

type Vendor struct {
	// List of non-default known models
	Models []*Model

	// Fallback driver
	DefaultConstructor func() (*Driver, error)
}


type DriverFactory struct {
	// map of all verndor drivers
	KnownDrivers map[string] *Vendor
}

func NewDriverFactory() *DriverFactory {
	return &DriverFactory{KnownDrivers:  map[string] *Vendor{}, }
}

func (df *DriverFactory) Register(v string, m string, c func() (*Driver, error)) error {
	if v == "" {
		v = "default"
	}

	if m == "" {
		m = "default"
	}

	if v == "default" && m != "default" {
		return fmt.Errorf("can't allow to register non-default model for default vendor")
	}

	var re *regexp.Regexp
	if m != "default" {
		re = regexp.MustCompile(m)
	}

	vendor, ok := df.KnownDrivers[v]
	if !ok {
		df.KnownDrivers[v] = &Vendor{}
	}
	vendor = df.KnownDrivers[v]

	if m == "default" {
		if vendor.DefaultConstructor != nil {
			return fmt.Errorf("trying to override default model for vendor %s", v)
		}
		vendor.DefaultConstructor = c
		return nil
	}

	vendor.Models = append(vendor.Models, &Model{Re: re, Constructor: c})
	return nil
}

func (df *DriverFactory) NewDriver(v string, m string) (*Driver, error) {
	if v == "" {
		v = "default"
	}

	if m == "" {
		m = "default"
	}

	if v == "default" && m != "default" {
		return nil, fmt.Errorf("model %s can't be identified without specifying vendor", m)
	}

	vendor, ok := df.KnownDrivers[v]
	if !ok {
		return nil, fmt.Errorf("can't find any driver for vendor %s", v)
	}

	if m == "default" {
		if vendor.DefaultConstructor == nil {
			return nil, fmt.Errorf("there is no default model registered in vendor %s drivers", v)
		}
		return vendor.DefaultConstructor()
	}

	// check in registration order
	for _, model := range vendor.Models {
		if model.Re.MatchString(m) {
			return model.Constructor()
		}
	}

	if vendor.DefaultConstructor != nil {
		return vendor.DefaultConstructor()
	}

	return nil, fmt.Errorf("wasn't able to find driver for %s %s", v, m)
}
