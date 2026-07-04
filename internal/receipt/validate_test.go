package receipt

import "testing"

func TestValidateReceiptRequiresMainFields(t *testing.T) {
	r := Receipt{}
	if err := Validate(r); err == nil {
		t.Fatal("Validate() returned nil for empty receipt")
	}
}
