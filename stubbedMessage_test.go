package runamqp

import "testing"

func TestStubMessage_Body(t *testing.T) {
	msg := NewStubMessage("some message")

	if string(msg.Body()) != "some message" {
		t.Error("body dont werk")
	}
}

func Test_It_Should_Ack_Message_Not_Acked_Previously(t *testing.T) {

	msg := NewStubMessage("msg")

	if msg.AckCalled() {
		t.Error("Ack should NOT have been called at this point")
	}

	err := msg.Ack()

	if err != nil {
		t.Error("Should have been able to successfully Ack the message", err)
	}

	if !msg.AckCalled() {
		t.Error("Message should have been Acked but it was not")
	}
}

func Test_It_Should_NOT_Ack_Message_Previously_Acked(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.AckCalled() {
		t.Error("Ack should NOT have been called at this point")
	}

	err := msg.Ack()

	if err != nil {
		t.Error("Should have been able to successfully Ack the message", err)
	}

	if !msg.AckCalled() {
		t.Error("Expect ack count to be 1")
	}

	err = msg.Ack()

	if err == nil {
		t.Error("Should NOT have been able to Ack a message that has been Acked previously")
	}
}

func Test_It_Should_NOT_Ack_Message_Previously_Nacked(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.AckCalled() {
		t.Error("Ack should NOT have been called at this point")
	}

	err := msg.Nack("Nacking reason")

	if err != nil {
		t.Error("Should have been able to successfully Nack the message", err)
	}

	if msg.AckCalled() {
		t.Error("Ack should NOT have been called at this point")
	}

	err = msg.Ack()

	if err == nil {
		t.Error("Should NOT have been able to Ack a message that has been Nacked previously")
	}
}

func Test_It_Should_NOT_Ack_Message_Previously_Requeued(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.AckCalled() {
		t.Error("Ack should NOT have been called at this point")
	}

	err := msg.Requeue("Requeuing reason")

	if err != nil {
		t.Error("Should have been able to successfully Nack the message", err)
	}

	if msg.AckCalled() {
		t.Error("Ack should NOT have been called at this point")
	}

	err = msg.Ack()

	if err == nil {
		t.Error("Should NOT have been able to Ack a message that has been Reqeued previously")
	}
}

func Test_It_Should_Nack_Message_Not_Nacked_Previosly(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.NackCalled() {
		t.Error("Nack should NOT have been called at this point")
	}

	err := msg.Nack("successful nack")

	if err != nil {
		t.Error("Should have been able to successfully Nack the message", err)
	}

	if !msg.NackCalled() {
		t.Error("Nack should have been called at this point")
	}

	if !msg.NackedWith("successful nack") {
		t.Error("Expected nack reason to be recorded")
	}
}

func Test_It_Should_NOT_Nack_Message_Previously_Nacked(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.NackCalled() {
		t.Error("Nack should NOT have been called at this point")
	}

	err := msg.Nack("successful nack")

	if err != nil {
		t.Error("Should have been able to successfully Nack the message", err)
	}

	if !msg.NackCalled() {
		t.Error("Nack should have been called at this point")
	}

	if !msg.NackedWith("successful nack") {
		t.Error("Expected nack reason to be recorded")
	}

	err = msg.Nack("unsuccessful nack")

	if err == nil {
		t.Error("Should NOT have been able to call the Nack when it was Nacked previously", err)
	}
}

func Test_It_Should_NOT_Nack_Message_Previously_Acked(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.NackCalled() {
		t.Error("Nack should NOT have been called at this point")
	}

	err := msg.Ack()

	if err != nil {
		t.Error("Should have been able to successfully Ack the message", err)
	}

	if msg.NackCalled() {
		t.Error("Nack should NOT have been called at this point")
	}

	err = msg.Nack("unsuccessful nack")

	if err == nil {
		t.Error("Should NOT have been able to call the Nack when it was Acked previously")
	}
}

func Test_It_Should_NOT_Nack_Message_Previously_Requeued(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.NackCalled() {
		t.Error("Nack should NOT have been called at this point")
	}

	err := msg.Requeue("successful nack")

	if err != nil {
		t.Error("Should have been able to successfully Requeue the message", err)
	}

	if msg.NackCalled() {
		t.Error("Nack should NOT have been called at this point")
	}

	err = msg.Nack("unsuccessful nack")

	if err == nil {
		t.Error("Should NOT have been able to call the Nack when it was Nacked previously")
	}
}

func Test_It_Should_Requeue_Message_Not_Requeued_Previously(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.RequeueCalled() {
		t.Error("Requeue should NOT have been called at this point")
	}

	err := msg.Requeue("successful requeue")

	if err != nil {
		t.Error("Should have been able to successfully Requeue the message", err)
	}

	if !msg.RequeueCalled() {
		t.Error("Requeue should have been called at this point")
	}

	if !msg.RequeuedWith("successful requeue") {
		t.Error("Expected requeue to been called with poo but it was not called.")
	}
}

func Test_It_Should_NOT_Requeue_Message_Previously_Requeued(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.RequeueCalled() {
		t.Error("Requeue should NOT have been called at this point")
	}

	err := msg.Requeue("successful requeue")

	if err != nil {
		t.Error("Should have been able to successfully Requeue the message", err)
	}

	if !msg.RequeueCalled() {
		t.Error("Requeue should have been called at this point")
	}

	if !msg.RequeuedWith("successful requeue") {
		t.Error("Expected requeue to been called with poo but it was not called.")
	}

	err = msg.Requeue("unsuccessful requeue")

	if err == nil {
		t.Error("Should NOT have been able to Requeue the message when it was Requeued previously")
	}
}

func Test_It_Should_NOT_Requeue_Message_Previously_Acked(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.RequeueCalled() {
		t.Error("Requeue should NOT have been called at this point")
	}

	err := msg.Ack()

	if err != nil {
		t.Error("Should have been able to successfully Ack the message", err)
	}

	if msg.RequeueCalled() {
		t.Error("Requeue should NOT have been called at this point")
	}

	err = msg.Requeue("unsuccessful requeue")

	if err == nil {
		t.Error("Should NOT have been able to Requeue the message when it was Acked previously")
	}
}

func Test_It_Should_NOT_Requeue_Message_Previously_Nacked(t *testing.T) {
	msg := NewStubMessage("msg")

	if msg.RequeueCalled() {
		t.Error("Requeue should NOT have been called at this point")
	}

	err := msg.Nack("successful nack")

	if err != nil {
		t.Error("Should have been able to successfully Nack the message", err)
	}

	if msg.RequeueCalled() {
		t.Error("Requeue should NOT have been called at this point")
	}

	err = msg.Requeue("unsuccessful requeue")

	if err == nil {
		t.Error("Should NOT have been able to Requeue the message when it was Nacked previously")
	}
}
