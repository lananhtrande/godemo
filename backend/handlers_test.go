package main

import (
	"bufio"
	"encoding/json"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func testStore() *Store {
	return NewStore([]Site{
		{ID: "a", Name: "Alpha", CapacityMWh: 10, MaxPowerMW: 5, SoC: 50, Status: "charging"},
		{ID: "b", Name: "Beta", CapacityMWh: 10, MaxPowerMW: 5, SoC: 50, Status: "discharging"},
	})
}

func TestListSites(t *testing.T) {
	srv := httptest.NewServer((&API{store: testStore(), allowedOrigin: "*"}).routes())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/sites")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", res.StatusCode)
	}
	var sites []Site
	if err := json.NewDecoder(res.Body).Decode(&sites); err != nil {
		t.Fatal(err)
	}
	if len(sites) != 2 || sites[0].Name != "Alpha" {
		t.Fatalf("unexpected sites: %+v", sites)
	}
}

func TestGetUnknownSiteReturns404(t *testing.T) {
	srv := httptest.NewServer((&API{store: testStore(), allowedOrigin: "*"}).routes())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/sites/does-not-exist")
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if res.StatusCode != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", res.StatusCode)
	}
}

func TestStreamSendsUpdates(t *testing.T) {
	store := testStore()
	srv := httptest.NewServer((&API{store: store, allowedOrigin: "*"}).routes())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/v1/sites/a/stream")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if ct := res.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Fatalf("content-type = %q", ct)
	}

	// Trigger one simulation step after the client is connected.
	go func() {
		time.Sleep(100 * time.Millisecond)
		store.Step(time.Minute, rand.New(rand.NewSource(1)))
	}()

	reader := bufio.NewReader(res.Body)
	events := 0
	deadline := time.Now().Add(2 * time.Second)
	for events < 2 && time.Now().Before(deadline) {
		line, err := reader.ReadString('\n')
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(line, "data: ") {
			var s Site
			if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &s); err != nil {
				t.Fatalf("bad event: %v", err)
			}
			if s.ID != "a" {
				t.Fatalf("got event for %q, want a", s.ID)
			}
			events++
		}
	}
	if events < 2 {
		t.Fatalf("got %d events, want initial state + 1 update", events)
	}
}

// The simulation must never leave the physical limits, whatever the random numbers are.
func TestSimulationKeepsSoCInRange(t *testing.T) {
	store := testStore()
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10_000; i++ {
		store.Step(time.Minute, rng)
	}
	for _, s := range store.List() {
		if s.SoC < 0 || s.SoC > 100 {
			t.Fatalf("%s: SoC %.2f out of range", s.ID, s.SoC)
		}
		if s.PowerMW > s.MaxPowerMW || s.PowerMW < -s.MaxPowerMW {
			t.Fatalf("%s: power %.2f exceeds max %.2f", s.ID, s.PowerMW, s.MaxPowerMW)
		}
	}
}

// Unsubscribing must remove the channel, so closed connections don't leak.
func TestUnsubscribeRemovesSubscriber(t *testing.T) {
	store := testStore()
	_, unsubscribe := store.Subscribe("a")
	unsubscribe()
	unsubscribe() // calling twice must be safe
	store.mu.RLock()
	defer store.mu.RUnlock()
	if n := len(store.subscribers["a"]); n != 0 {
		t.Fatalf("subscribers = %d, want 0", n)
	}
}
