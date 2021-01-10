package raw

import "testing"

func TestGetID(t *testing.T) {
	h := make(Header, HeaderSize)
	expected := 44
	h.Encode(INIT, 10, uint32(expected))
	got := h.ID()
	if uint32(expected) != got {
		t.Errorf("Expected id to be %d but instead got %d", expected, got)
	}
}

func TestProtoVersion(t *testing.T) {
	h := make(Header, HeaderSize)
	expected := ProtoVersion
	h.Encode(INIT, 10, 44)
	got := h.ProtoVersion()
	if uint16(expected) != got {
		t.Errorf("Expected proto version to be %d but instead got %d", expected, got)
	}
}

func TestNext(t *testing.T) {
	h := make(Header, HeaderSize)
	expected := 10
	h.Encode(INIT, uint32(expected), 44)
	got := h.Next()
	if uint32(expected) != got {
		t.Errorf("Expected next payload size to be %d but instead got %d", expected, got)
	}
}

func TestMessageType(t *testing.T) {
	h := make(Header, HeaderSize)
	expected := INIT
	h.Encode(expected, 10, 44)
	got := h.MessageType()
	if expected != got {
		t.Errorf("Expected message type to be %d but instead got %d", expected, got)
	}
}
