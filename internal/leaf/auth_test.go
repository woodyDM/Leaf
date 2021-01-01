package leaf

import "testing"

func TestEnc(t *testing.T) {
	pass := encPass("123456", "1234")
	if pass == "" {
		t.Error("no pass")
	}
}

func TestNameMatch(t *testing.T) {
	if !nameMatch("1ab2"){
		t.Error("e1")
	}
	if !nameMatch("a"){
		t.Error("e1")
	}
	if !nameMatch("b"){
		t.Error("e1")
	}
	if !nameMatch("c"){
		t.Error("e1")
	}
	if nameMatch(""){
		t.Error("e1")
	}
	if nameMatch("c -["){
		t.Error("e last")
	}
}


