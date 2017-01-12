package runamqp

import "testing"

func TestReturnsUnrecognisedTypeAndErrorWhenExchangeIsWrong(t *testing.T) {
	typ, err := NewExchangeType("cj rules")

	if err == nil {
		t.Error("We expect an error for this type")
	}

	if typ != Unrecognised {
		t.Error("Expected unrecognised type but got", typ)
	}
}

func TestItsCaseInsensitiveWhenCheckingExchangeTypes(t *testing.T) {
	typ, err := NewExchangeType("fANout")

	if err != nil {
		t.Fatal("Unexpected error", err)
	}

	if typ != Fanout {
		t.Error("Unexpected type, expected Fanout but got", typ)
	}
}
