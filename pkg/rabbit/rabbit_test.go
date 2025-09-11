package rabbit

import "testing"

func SendTest(t *testing.T) {
	ok := Send("test", map[string]string{"hello": "world"})
	if !ok {
		t.Error("Failed to send message in test mode")
	}
}
