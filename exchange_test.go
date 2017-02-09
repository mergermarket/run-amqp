package runamqp

import (
	"github.com/golang/mock/gomock"
	"github.com/mergermarket/run-amqp/connection"
	"testing"
)

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

func TestMakeExchange(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	t.Run("unrecognised exchange type returns error and no exchanges declared", func(t *testing.T) {
		ch := connection.NewMockAMQPChannel(mockCtrl)

		exchangeName := "whatever"

		err := makeExchange(ch, exchangeName, Unrecognised)

		if err == nil {
			t.Fatal("Expected an error due to unrecognised exchange type")
		}

		ch.EXPECT().ExchangeDeclare(exchangeName,
			string(Unrecognised),
			true,
			false,
			false,
			false,
			nil).Return(nil).Times(0)
	})

	t.Run("makes an exchange with the configured exchange name and type with durable true, not auto-deleting, not internal and no wait", func(t *testing.T) {
		ch := connection.NewMockAMQPChannel(mockCtrl)

		exchangeName := "whatever"

		ch.EXPECT().ExchangeDeclare(exchangeName,
			string(Fanout),
			true,
			false,
			false,
			false,
			nil).Return(nil).Times(1)

		err := makeExchange(ch, exchangeName, Fanout)

		if err != nil {
			t.Fatal("Dont expect an error")
		}
	})

}
