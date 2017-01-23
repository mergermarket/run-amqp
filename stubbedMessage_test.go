package runamqp

import "testing"

func TestItRecordsAcks(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.Calls.AckCount != 0 {
		t.Error("Expect ack count to be zero at the start")
	}

	msg.Ack()

	if msg.Calls.AckCount != 1 {
		t.Error("Expect ack count to be 1")
	}

	if !msg.AckedOnce() {
		t.Error("Expected to be acked once")
	}

	msg.Ack()

	if msg.AckedOnce() {
		t.Error("Dont expect to be acked once anymore")
	}
}

func TestItRecordsNacks(t *testing.T) {
	msg := NewStubMessage("msg")

	if len(msg.Calls.Nack) != 0 {
		t.Error("Expect nack count to be zero at the start")
	}

	msg.Nack("poo")

	if len(msg.Calls.Nack) != 1 {
		t.Error("Expect nack count to be 1")
	}

	if msg.Calls.Nack[0] != "poo" {
		t.Error("Expected nack reason to be recorded")
	}

	msg.Nack("butt")

	if len(msg.Calls.Nack) != 2 {
		t.Error("expected two calls to nack")
	}

	if msg.Calls.Nack[1] != "butt" {
		t.Error("expected nack call to be recorded")
	}
}

func TestStubMessage_Body(t *testing.T) {
	msg := NewStubMessage("some message")

	if string(msg.Body()) != "some message" {
		t.Error("body dont werk")
	}
}

func TestItRecordsNaqueue(t *testing.T) {
	msg := NewStubMessage("msg")

	if len(msg.Calls.Requeue) != 0 {
		t.Error("Expect Requeue count to be zero at the start")
	}

	msg.Requeue("poo")

	if len(msg.Calls.Requeue) != 1 {
		t.Error("Expect Requeue count to be 1")
	}

	if msg.Calls.Requeue[0] != "poo" {
		t.Error("Expected Requeue reason to be recorded")
	}

	msg.Requeue("butt")

	if len(msg.Calls.Requeue) != 2 {
		t.Error("expected two calls to Requeue")
	}

	if msg.Calls.Requeue[1] != "butt" {
		t.Error("expected Requeue call to be recorded")
	}
}
