package crypter

import (
	"testing"

	//"log"
)

func TestEncryptDecrypt(t *testing.T) {
	k, err := key("password")
	if err != nil {
		t.Errorf("key failed: %v", err)
	}

	value := "some text"
	ve, err := encrypt(value, k)
	if err != nil {
		t.Errorf("encrypt failed: %v", err)
	}

	//log.Printf("ve: %v", ve)

	v, err := decrypt(ve, k)
	if err != nil {
		t.Errorf("decrypt failed: %v", err)
	}

	if v != value {
		t.Errorf("expected %s, got %s", value, v)
	}
}
