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
<title>metro</title>
<style>
  *, *::before, *::after { box-sizing: border-box; }
  body {
    font-family: 'SF Mono', 'Cascadia Code', 'Fira Code', 'JetBrains Mono', 'Consolas', monospace;
    display: flex; justify-content: center; align-items: center;
    height: 100vh; margin: 0;
    background: #0a0a0a; color: #b0b0b0;
    font-size: 13px; line-height: 1.6;
  }
  .card {
    border: 1px solid #222; background: #141414;
    max-width: 440px; width: 90%;
  }
  .header {
    padding: 10px 16px; border-bottom: 1px solid #222;
    color: #555; font-size: 12px;
    display: flex; align-items: center; gap: 8px;
  }
  .header span { color: #67e8f9; }
  .body { padding: 24px; }
  .line { display: flex; align-items: center; gap: 8px; margin: 6px 0; }
  .line .dim { color: #444; }
  .cursor {
    display: inline-block; width: 7px; height: 14px;
    background: #67e8f9; animation: blink 1s step-end infinite;
    vertical-align: middle; margin-left: 2px;
  }
  @keyframes blink { 50% { opacity: 0; } }
  .ok { color: #4ade80; }
  .err { color: #f87171; }
  .dim { color: #555; }
  .label { color: #888; margin: 16px 0 8px; }
  input[type="text"], input[type="number"] {
    background: #1a1a1a; border: 1px solid #333; color: #c0c0c0;
    padding: 8px 12px; width: 100%; margin: 4px 0;
    font-family: inherit; font-size: 13px; outline: none;
    transition: border-color 0.15s;
  }
  input:focus { border-color: #67e8f9; }
  input::placeholder { color: #444; }
  button {
    background: transparent; border: 1px solid #67e8f9; color: #67e8f9;
    padding: 7px 20px; cursor: pointer; font-family: inherit;
    font-size: 12px; margin-top: 8px; transition: all 0.15s;
  }
  button:hover { background: rgba(103,232,249,0.1); }
  details { margin-top: 20px; }
  summary {
    color: #555; cursor: pointer; font-size: 12px;
    list-style: none; user-select: none;
  }
  summary::-webkit-details-marker { display: none; }
  summary::before { content: '\25b8  '; }
  details[open] summary::before { content: '\25be  '; }
  .coords { margin-top: 12px; }
  #search-status { margin-top: 8px; font-size: 12px; min-height: 18px; }
  .fadein { animation: fadein 0.2s ease; }
  @keyframes fadein { from { opacity: 0; } to { opacity: 1; } }
</style>
</head>
<body>
<div class="card">
  <div class="header"><span>&gt;</span> metro location</div>
  <div class="body fadein" id="content">
    <div class="line"><span class="dim">~</span> requesting permission<span class="cursor"></span></div>
    <div class="line dim" style="font-size:12px">waiting for browser geolocation...</div>
  </div>
</div>
<script>
function sendCoords(lat, lon) {
  fetch('/callback', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ lat: lat, lon: lon })
  }).then(function() {
    document.getElementById('content').innerHTML =
      '<div class="line fadein"><span class="ok">[ok]</span> location shared &mdash; closing...</div>';
    setTimeout(function() { window.close(); }, 1500);
  });
}

function showManualForm(reason) {
  document.getElementById('content').innerHTML =
    '<div class="fadein">' +
    '<div class="line"><span class="err">[err]</span> ' + reason + '</div>' +
    '<p class="label">search for a station or address</p>' +
    '<input id="addr" type="text" placeholder="Place d\'Italie, Chatelet, 73 rue Rivoli..." autofocus>' +
    '<button onclick="searchAddress()">search</button>' +
    '<div id="search-status"></div>' +
    '<details><summary>enter coordinates manually</summary>' +
    '<div class="coords">' +
    '<input id="lat" type="number" step="any" placeholder="latitude &mdash; e.g. 48.8566">' +
    '<input id="lon" type="number" step="any" placeholder="longitude &mdash; e.g. 2.3522">' +
    '<button onclick="submitManual()">share location</button>' +
    '</div></details>' +
    '</div>';
  document.getElementById('addr').addEventListener('keydown', function(e) {
    if (e.key === 'Enter') searchAddress();
  });
}

function searchAddress() {
  var q = document.getElementById('addr').value.trim();
  if (!q) return;
  var status = document.getElementById('search-status');
  status.textContent = 'searching...';
  status.className = 'dim';
  fetch('https://nominatim.openstreetmap.org/search?format=json&limit=1&countrycodes=fr&q=' + encodeURIComponent(q))
    .then(function(r) { return r.json(); })
    .then(function(data) {
      if (!data.length) { status.textContent = 'no results found'; status.className = 'err'; return; }
      var lat = parseFloat(data[0].lat), lon = parseFloat(data[0].lon);
      status.innerHTML = '<span class="ok">[ok]</span> ' + data[0].display_name.split(',').slice(0,2).join(',');
      status.className = '';
      setTimeout(function() { sendCoords(lat, lon); }, 500);
    })
    .catch(function() { status.textContent = 'search failed \u2014 try coordinates instead'; status.className = 'err'; });
}

function submitManual() {
  var lat = parseFloat(document.getElementById('lat').value);
  var lon = parseFloat(document.getElementById('lon').value);
  if (isNaN(lat) || isNaN(lon)) { alert('enter valid coordinates'); return; }
  sendCoords(lat, lon);
}

if (!navigator.geolocation) {
  showManualForm('geolocation unavailable (requires HTTPS or localhost)');
} else {
  navigator.geolocation.getCurrentPosition(
    function(pos) { sendCoords(pos.coords.latitude, pos.coords.longitude); },
    function(err) { showManualForm(err.message.toLowerCase()); },
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
func GetLocation(timeout time.Duration, port int) (lat, lon float64, err error) {
	return getLocation(timeout, port)
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
	localhostURL := fmt.Sprintf("http://localhost:%d/locate", port)
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
