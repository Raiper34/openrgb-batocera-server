package main

import (
	"embed"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	sdk "github.com/csutorasa/go-openrgb-sdk"

	"github.com/raiper34/openrgb-batocera-server/internal/api"
	"github.com/raiper34/openrgb-batocera-server/internal/openrgb"
	"github.com/raiper34/openrgb-batocera-server/internal/state"
)

//go:embed web
var webFiles embed.FS

func main() {
	port        := flag.Int("port", 8080, "HTTP server port")
	openrgbHost := flag.String("openrgb-host", "", "Auto-connect to OpenRGB host on startup (e.g. localhost)")
	openrgbPort := flag.Int("openrgb-port", 6742, "OpenRGB server port (used with --openrgb-host)")
	stateFile   := flag.String("state-file", "state.json", "Path to the persistent state file")
	flag.Parse()

	appState := state.New(*stateFile)
	manager  := openrgb.NewManager()

	// Determine connection target: CLI flag > OPENRGB_HOST env var > saved state
	envHost    := os.Getenv("OPENRGB_HOST")
	envPortStr := os.Getenv("OPENRGB_PORT")

	host     := *openrgbHost
	port6742 := *openrgbPort

	if host == "" && envHost != "" {
		host = envHost
		if envPortStr != "" {
			if p, err := strconv.Atoi(envPortStr); err == nil {
				port6742 = p
			}
		}
		log.Printf("Using OPENRGB_HOST env var: %s:%d", host, port6742)
	}

	if host == "" {
		if saved := appState.GetConnection(); saved != nil {
			host     = saved.Host
			port6742 = saved.Port
			log.Printf("Loaded saved connection from state: %s:%d", host, port6742)
		}
	}

	if host != "" {
		log.Printf("Auto-connecting to OpenRGB at %s:%d ...", host, port6742)
		if err := manager.Connect(host, port6742); err != nil {
			log.Printf("Auto-connect failed: %v", err)
		} else {
			log.Printf("Auto-connect successful.")
			appState.SetConnection(host, port6742)
			restoreState(manager, appState)
		}
	}

	handler := api.NewHandler(manager, appState, envHost)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Serve Angular frontend (SPA with fallback to index.html)
	webFS, err := fs.Sub(webFiles, "web")
	if err != nil {
		log.Fatalf("Failed to create web file system: %v", err)
	}

	fileServer := http.FileServer(http.FS(webFS))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")

		f, err := webFS.Open(path)
		if err != nil {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		defer f.Close()

		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}

		fileServer.ServeHTTP(w, r)
	})

	finalHandler := corsMiddleware(loggingMiddleware(mux))

	addr := fmt.Sprintf(":%d", *port)
	log.Printf("OpenRGB Batocera Server starting on http://0.0.0.0%s", addr)
	log.Printf("Web UI: http://localhost%s", addr)
	log.Printf("API:    http://localhost%s/api/", addr)

	if err := http.ListenAndServe(addr, finalHandler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

// restoreState applies persisted device colors and modes after a successful connection.
func restoreState(manager *openrgb.Manager, appState *state.State) {
	saved := appState.GetDevices()
	if len(saved) == 0 {
		return
	}
	log.Printf("Restoring state for %d device(s)...", len(saved))

	for _, ds := range saved {
		idx := uint32(ds.Index)

		// Restore mode
		if ds.ModeID > 0 {
			device, err := manager.GetDevice(idx)
			if err == nil {
				if err := manager.SetMode(idx, int32(ds.ModeID), device); err != nil {
					log.Printf("  device %d: set mode failed: %v", idx, err)
				} else {
					log.Printf("  device %d (%s): mode restored to %d", idx, ds.Name, ds.ModeID)
				}
			}
		}

		// Restore colors
		if len(ds.Colors) > 0 {
			colors := make([]sdk.Color, len(ds.Colors))
			for i, hex := range ds.Colors {
				colors[i] = hexToColor(hex)
			}
			if err := manager.SetDeviceColors(idx, colors); err != nil {
				log.Printf("  device %d: set colors failed: %v", idx, err)
			} else {
				log.Printf("  device %d (%s): %d colors restored", idx, ds.Name, len(colors))
			}
		}
	}
}

// hexToColor parses a 6-char hex string "RRGGBB" into an sdk.Color.
func hexToColor(hex string) sdk.Color {
	if len(hex) != 6 {
		return sdk.Color{}
	}
	parse := func(s string) uint8 {
		var v uint8
		fmt.Sscanf(s, "%02X", &v)
		return v
	}
	return sdk.Color{
		R: parse(hex[0:2]),
		G: parse(hex[2:4]),
		B: parse(hex[4:6]),
	}
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			log.Printf("%s %s", r.Method, r.URL.Path)
		}
		next.ServeHTTP(w, r)
	})
}
