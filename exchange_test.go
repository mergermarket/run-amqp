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

	tests := []struct {
		input    string
		expected ExchangeType
	}{
		{input: "fanOUT", expected: Fanout},
		{input: "fanout", expected: Fanout},
		{input: "topic", expected: Topic},
		{input: "toPic", expected: Topic},
		{input: "direct", expected: Direct},
		{input: "Direct", expected: Direct},
	}

	for _, tst := range tests {
		typ, err := NewExchangeType(tst.input)

		if err != nil {
			t.Fatal("Unexpected error", err)
		}

		if typ != tst.expected {
			t.Error("Unexpected type, expected Fanout but got", typ)
		}
	}

}
