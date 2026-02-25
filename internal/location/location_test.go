package location

import (
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestCallback(t *testing.T) {
	const port = 18923
	done := make(chan struct{})
	var lat, lon float64
	var locErr error

	go func() {
		lat, lon, locErr = getLocation(10*time.Second, port)
		close(done)
	}()

	// Wait for the server to be ready
	time.Sleep(200 * time.Millisecond)

	// Simulate a browser posting coordinates
	resp, err := http.Post(
		"http://127.0.0.1:18923/callback",
		"application/json",
		strings.NewReader(`{"lat":48.860611,"lon":2.337644}`),
	)
	if err != nil {
		t.Fatalf("POST /callback failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	<-done
	if locErr != nil {
		t.Fatalf("getLocation returned error: %v", locErr)
	}
	if lat < 48.86 || lat > 48.87 {
		t.Errorf("unexpected lat: %f", lat)
	}
	if lon < 2.33 || lon > 2.34 {
		t.Errorf("unexpected lon: %f", lon)
	}
	t.Logf("lat=%.6f lon=%.6f", lat, lon)
}

func TestServesHTML(t *testing.T) {
	const port = 18924
	go func() {
		getLocation(5*time.Second, port)
	}()

	time.Sleep(200 * time.Millisecond)

	resp, err := http.Get("http://127.0.0.1:18924/locate")
	if err != nil {
		t.Fatalf("GET /locate failed: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("expected text/html, got %s", ct)
	}
}
