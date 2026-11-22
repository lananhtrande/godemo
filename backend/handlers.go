package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
)

type API struct {
	store         *Store
	allowedOrigin string
}

// routes uses the method + path patterns of Go 1.22+ (no framework needed).
func (a *API) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", a.health)
	mux.HandleFunc("GET /api/v1/sites", a.listSites)
	mux.HandleFunc("GET /api/v1/sites/{id}", a.getSite)
	mux.HandleFunc("GET /api/v1/sites/{id}/stream", a.streamSite)
	return a.cors(logRequests(mux))
}

func (a *API) health(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) listSites(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, a.store.List())
}

func (a *API) getSite(w http.ResponseWriter, r *http.Request) {
	site, ok := a.store.Get(r.PathValue("id"))
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "site not found"})
		return
	}
	writeJSON(w, http.StatusOK, site)
}

// streamSite pushes every update of one site to the browser using
// Server-Sent Events (SSE): one long-lived HTTP response, one "data:" line per event.
func (a *API) streamSite(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	site, ok := a.store.Get(id)
	if !ok {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "site not found"})
		return
	}
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // tell reverse proxies (nginx) not to buffer

	updates, unsubscribe := a.store.Subscribe(id)
	defer unsubscribe() // runs when the client disconnects -> no goroutine/channel leak

	send := func(s Site) error {
		b, err := json.Marshal(s)
		if err != nil {
			return err
		}
		if _, err := fmt.Fprintf(w, "data: %s\n\n", b); err != nil {
			return err
		}
		flusher.Flush()
		return nil
	}

	// Send the current state immediately, so the UI doesn't wait for the next tick.
	if err := send(site); err != nil {
		return
	}

	heartbeat := time.NewTicker(15 * time.Second) // keeps idle connections alive through proxies
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done(): // browser closed the tab or navigated away
			return
		case s, ok := <-updates:
			if !ok {
				return
			}
			if err := send(s); err != nil {
				return
			}
		case <-heartbeat.C:
			fmt.Fprint(w, ": ping\n\n") // SSE comment line, ignored by the browser
			flusher.Flush()
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("write json: %v", err)
	}
}

// cors lets the Next.js app (another origin) call this API from the browser.
func (a *API) cors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", a.allowedOrigin)
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}
