package location

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"time"
)

type coords struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

const page = `<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width">
<title>metro - location</title>
<style>
  body { font-family: -apple-system, system-ui, sans-serif; display: flex;
         justify-content: center; align-items: center; height: 100vh; margin: 0;
         background: #1a1a2e; color: #e0e0e0; }
  .box { text-align: center; max-width: 400px; padding: 20px; }
  .spinner { display: inline-block; width: 24px; height: 24px;
             border: 3px solid #444; border-top-color: #0af; border-radius: 50%;
             animation: spin 0.8s linear infinite; margin-bottom: 12px; }
  @keyframes spin { to { transform: rotate(360deg); } }
  .done { color: #0f6; }
  .err  { color: #f44; }
  input { background: #2a2a3e; border: 1px solid #444; color: #e0e0e0;
          padding: 8px 12px; border-radius: 4px; width: 200px; margin: 4px 0; }
  button { background: #0af; color: #1a1a2e; border: none; padding: 8px 20px;
           border-radius: 4px; cursor: pointer; font-weight: bold; margin-top: 8px; }
  button:hover { background: #09e; }
  .hint { font-size: 0.85em; color: #888; margin-top: 12px; }
</style>
</head>
<body>
<div class="box" id="content">
  <div class="spinner"></div>
  <p>Sharing your location with <strong>metro</strong> CLI...</p>
</div>
<script>
function sendCoords(lat, lon) {
  fetch('/callback', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lat: lat, lon: lon })
  }).then(function() {
    document.getElementById('content').innerHTML =
      '<p class="done">Location shared. You can close this tab.</p>';
    setTimeout(function() { window.close(); }, 1500);
  });
}

function showManualForm(reason) {
  document.getElementById('content').innerHTML =
    '<p class="err">' + reason + '</p>' +
    '<p>Search for a station or address:</p>' +
    '<input id="addr" type="text" placeholder="e.g. Place d\'Italie" style="width:240px" autofocus><br>' +
    '<button onclick="searchAddress()">Search</button>' +
    '<p id="search-status" class="hint"></p>' +
    '<details style="margin-top:16px"><summary class="hint" style="cursor:pointer">Or enter coordinates manually</summary>' +
    '<div style="margin-top:8px">' +
    '<input id="lat" type="number" step="any" placeholder="Latitude (e.g. 48.8566)"><br>' +
    '<input id="lon" type="number" step="any" placeholder="Longitude (e.g. 2.3522)"><br>' +
    '<button onclick="submitManual()">Share location</button>' +
    '</div></details>';
  document.getElementById('addr').addEventListener('keydown', function(e) {
    if (e.key === 'Enter') searchAddress();
  });
}

function searchAddress() {
  var q = document.getElementById('addr').value.trim();
  if (!q) return;
  var status = document.getElementById('search-status');
  status.textContent = 'Searching...';
  status.className = 'hint';
  fetch('https://nominatim.openstreetmap.org/search?format=json&limit=1&countrycodes=fr&q=' + encodeURIComponent(q))
    .then(function(r) { return r.json(); })
    .then(function(data) {
      if (!data.length) { status.textContent = 'No results found. Try a different query.'; status.className = 'err'; return; }
      var lat = parseFloat(data[0].lat), lon = parseFloat(data[0].lon);
      status.textContent = 'Found: ' + data[0].display_name.split(',').slice(0,2).join(',');
      status.className = 'done';
      setTimeout(function() { sendCoords(lat, lon); }, 500);
    })
    .catch(function() { status.textContent = 'Search failed. Try coordinates instead.'; status.className = 'err'; });
}

function submitManual() {
  var lat = parseFloat(document.getElementById('lat').value);
  var lon = parseFloat(document.getElementById('lon').value);
  if (isNaN(lat) || isNaN(lon)) { alert('Please enter valid coordinates'); return; }
  sendCoords(lat, lon);
}

if (!navigator.geolocation) {
  showManualForm('Geolocation is not available (requires HTTPS or localhost).');
} else {
  navigator.geolocation.getCurrentPosition(
    function(pos) { sendCoords(pos.coords.latitude, pos.coords.longitude); },
    function(err) { showManualForm('Could not get location: ' + err.message); },
    { enableHighAccuracy: true, timeout: 15000, maximumAge: 0 }
  );
}
</script>
</body>
</html>`

// GetLocation opens a browser to capture the user's GPS coordinates via the
// browser's Geolocation API. Works on macOS, Linux, and Windows.
// If geolocation is blocked (e.g. non-HTTPS access from LAN), the page shows
// a manual coordinate input form as fallback.
func GetLocation(timeout time.Duration) (lat, lon float64, err error) {
	return getLocation(timeout, 0)
}

// getLocation is the internal implementation. If fixedPort > 0, it uses that
// port (for testing); otherwise it picks a random free port.
func getLocation(timeout time.Duration, fixedPort int) (float64, float64, error) {
	// Bind to all interfaces so the page is reachable from the LAN
	// (needed when running on a NAS or headless server).
	addr := "0.0.0.0:0"
	if fixedPort > 0 {
		addr = fmt.Sprintf("0.0.0.0:%d", fixedPort)
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return 0, 0, fmt.Errorf("starting location server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port

	result := make(chan coords, 1)
	errCh := make(chan error, 1)

	mux := http.NewServeMux()

	mux.HandleFunc("/locate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, page)
	})

	mux.HandleFunc("/callback", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}
		var c coords
		if err := json.NewDecoder(r.Body).Decode(&c); err != nil {
			http.Error(w, "bad payload", http.StatusBadRequest)
			errCh <- fmt.Errorf("decoding location: %w", err)
			return
		}
		w.WriteHeader(http.StatusOK)
		result <- c
	})

	srv := &http.Server{Handler: mux}

	go func() {
		if err := srv.Serve(listener); err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Build URLs — localhost for local use, LAN IP for remote access
	localhostURL := fmt.Sprintf("http://127.0.0.1:%d/locate", port)
	lanIP := localIP()

	_ = openBrowser(localhostURL) // best-effort, may fail on headless systems

	if lanIP != "127.0.0.1" {
		lanURL := fmt.Sprintf("http://%s:%d/locate", lanIP, port)
		fmt.Printf("Open in your browser: %s\n", lanURL)
		fmt.Printf("  (or locally: %s)\n", localhostURL)
	} else {
		fmt.Printf("Open in your browser: %s\n", localhostURL)
	}

	// Wait for result or timeout
	select {
	case c := <-result:
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		return c.Lat, c.Lon, nil
	case e := <-errCh:
		srv.Close()
		return 0, 0, e
	case <-time.After(timeout):
		srv.Close()
		return 0, 0, fmt.Errorf("timed out after %s (did you allow browser location access?)", timeout)
	}
}

// localIP returns the LAN IP address of the machine by checking which
// local address would be used to reach an external host.
func localIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "127.0.0.1"
	}
	defer conn.Close()
	return conn.LocalAddr().(*net.UDPAddr).IP.String()
}

func openBrowser(url string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}
	return cmd.Start()
}
