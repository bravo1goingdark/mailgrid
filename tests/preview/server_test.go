// preview/server_test.go
package preview

import (
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestStartServer(t *testing.T) {
	port := 9091
	shutdown := make(chan struct{})
	done := make(chan struct{})

	go func() {
		if err := StartServer("<h1>hi</h1>", port, shutdown); err != nil {
			t.Errorf("server error: %v", err)
		}
		close(done)
	}()

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d", port))
	if err != nil {
		t.Fatalf("GET failed: %v", err)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatalf("close body: %v", err)
	}
	if string(body) != "<h1>hi</h1>" {
		t.Errorf("unexpected body: %s", body)
	}

	// trigger shutdown
	close(shutdown)

	select {
	case <-done:
	case <-time.After(1 * time.Second):
		t.Fatal("server did not shut down")
	}
}
