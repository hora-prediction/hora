package rbridge

import (
	"testing"
)

func TestGetRSession(t *testing.T) {
	SetHostname("localhost")
	SetPort(6311)
	sessionA, err := GetRSession("A")
	if err != nil {
		t.Error("Error: ", err)
	}

	sessionA.Eval("a <- 2")
	sessionA.Eval("b <- 3")
	ret, err := sessionA.Eval("a+b")
	if err != nil {
		t.Error("Error: ", err)
	}
	if ret != 5.0 {
		t.Error("Expected 5 but got", ret)
	}

	SetHostname("localhost")
	SetPort(6311)
	sessionB, err := GetRSession("B")
	if err != nil {
		t.Error("Error: ", err)
	}

	ret, err = sessionB.Eval("a+b")
	if err == nil {
		t.Error("Expected non-nil error but got nil with return value", ret)
	}

	sessionB.Eval("a <- 2.5")
	sessionB.Eval("b <- 3.4")
	ret, err = sessionB.Eval("a+b")
	if err != nil {
		t.Error("Error: ", err)
	}
	if ret != 5.9 {
		t.Error("Expected 5.9 but got", ret)
	}

	ret, err = sessionA.Eval("a+b")
	if err != nil {
		t.Error("Error: ", err)
	}
	if ret != 5.0 {
		t.Error("Expected 5.0 but got", ret)
	}
}
