package location

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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
		lat, lon, locErr = getLocation(10*time.Second, port, false)
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
		getLocation(5*time.Second, port, false)
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

func TestSaveCacheAndLoadCache(t *testing.T) {
	// Use a temp dir to avoid touching the real cache
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	SaveCache(48.8566, 2.3522)

	// Verify file was written
	data, err := os.ReadFile(filepath.Join(tmp, ".metro_location_cache.json"))
	if err != nil {
		t.Fatalf("cache file not written: %v", err)
	}
	var c cachedLocation
	if err := json.Unmarshal(data, &c); err != nil {
		t.Fatalf("invalid cache JSON: %v", err)
	}
	if c.Lat != 48.8566 || c.Lon != 2.3522 {
		t.Errorf("cache coords = (%.4f, %.4f), want (48.8566, 2.3522)", c.Lat, c.Lon)
	}

	// Load with generous TTL — should succeed
	lat, lon, err := LoadCache(5 * time.Minute)
	if err != nil {
		t.Fatalf("LoadCache failed: %v", err)
	}
	if lat != 48.8566 || lon != 2.3522 {
		t.Errorf("LoadCache = (%.4f, %.4f), want (48.8566, 2.3522)", lat, lon)
	}
}

func TestLoadCacheExpired(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	// Write a cache entry with old timestamp
	old := cachedLocation{Lat: 48.0, Lon: 2.0, CachedAt: time.Now().Add(-10 * time.Minute)}
	data, _ := json.Marshal(old)
	os.WriteFile(filepath.Join(tmp, ".metro_location_cache.json"), data, 0644)

	_, _, err := LoadCache(5 * time.Minute)
	if err == nil {
		t.Error("expected error for expired cache")
	}
}

func TestLoadCacheMissing(t *testing.T) {
	tmp := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", tmp)
	defer os.Setenv("HOME", origHome)

	_, _, err := LoadCache(5 * time.Minute)
	if err == nil {
		t.Error("expected error when cache file doesn't exist")
	}
}
