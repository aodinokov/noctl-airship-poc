package crypter

import (
	"testing"
	//"log"
)

func TestEncryptDecrypt(t *testing.T) {
	k, err := Key("testpass")
	if err != nil {
		t.Errorf("key failed: %v", err)
	}

	value := "some text"
	ve, err := Encrypt(value, k)
	if err != nil {
		t.Errorf("encrypt failed: %v", err)
	}

	//log.Printf("k: %v, ve: %v", k, ve)

	v, err := Decrypt(ve, k)
	if err != nil {
		t.Errorf("decrypt failed: %v", err)
	}

	if v != value {
		t.Errorf("expected %s, got %s", value, v)
	}
}
