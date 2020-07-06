package redfish

import (
	"testing"
)

func TestDriverFactory(t *testing.T) {
	f := NewDriverFactory()
	if f == nil {
		t.Error("can't create DriverFactory")
	}

	called := false
	fn := func(_ *RedfishOperationFunction) (Driver, error) {
		called = true
		return nil, nil
	}

	if f.Register("default", "", fn) != nil {
		t.Error("expected that default vendor/module would be created")
	}

	rfn, err := f.GetCreateDriverFn("default", "")
	if err != nil {
		t.Errorf("expected that the registered driver would be found, but got error: %v", err)
	}
	if _, err := rfn(nil); err != nil {
		t.Errorf("func shouldn't return err")
	}

	if !called {
		t.Error("expected that the registered function would be called")
	}
}
