package main

import (
	"math/rand"
	"sort"
	"sync"
	"time"
)

// Site is the current state of one battery storage site.
// Power is in MW: positive = discharging to the grid, negative = charging.
type Site struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	CapacityMWh float64   `json:"capacityMWh"`
	MaxPowerMW  float64   `json:"maxPowerMW"`
	SoC         float64   `json:"soc"`
	PowerMW     float64   `json:"powerMW"`
	Status      string    `json:"status"` // charging | discharging | idle
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Store keeps the latest state of every site and fans out updates
// to all subscribers (one channel per open SSE connection).
type Store struct {
	mu          sync.RWMutex
	sites       map[string]*Site
	subscribers map[string]map[chan Site]struct{}
}

func NewStore(sites []Site) *Store {
	s := &Store{
		sites:       make(map[string]*Site),
		subscribers: make(map[string]map[chan Site]struct{}),
	}
	for i := range sites {
		site := sites[i]
		site.UpdatedAt = time.Now().UTC()
		s.sites[site.ID] = &site
	}
	return s
}

// List returns copies of all sites, sorted by name.
func (s *Store) List() []Site {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Site, 0, len(s.sites))
	for _, site := range s.sites {
		out = append(out, *site) // copy, so callers can't modify our state
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Get returns a copy of one site.
func (s *Store) Get(id string) (Site, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	site, ok := s.sites[id]
	if !ok {
		return Site{}, false
	}
	return *site, true
}

// Subscribe registers a channel that receives every update of one site.
// The returned function must be called to unsubscribe (e.g. when the client disconnects).
func (s *Store) Subscribe(id string) (<-chan Site, func()) {
	ch := make(chan Site, 8) // small buffer absorbs short hiccups of a slow client
	s.mu.Lock()
	if s.subscribers[id] == nil {
		s.subscribers[id] = make(map[chan Site]struct{})
	}
	s.subscribers[id][ch] = struct{}{}
	s.mu.Unlock()

	unsubscribe := func() {
		s.mu.Lock()
		defer s.mu.Unlock()
		if _, ok := s.subscribers[id][ch]; ok {
			delete(s.subscribers[id], ch)
			close(ch)
		}
	}
	return ch, unsubscribe
}

// Step advances the simulation by dt of "site time" and notifies subscribers.
func (s *Store) Step(dt time.Duration, rng *rand.Rand) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for id, site := range s.sites {
		simulate(site, dt, rng)
		site.UpdatedAt = now

		// Non-blocking send: a slow client must never block the simulation
		// (or, in a real system, the ingestion of data from other sites).
		for ch := range s.subscribers[id] {
			select {
			case ch <- *site:
			default: // buffer full -> drop this update for that client, it gets the next one
			}
		}
	}
}

// simulate applies a simple charge/discharge model to one site.
func simulate(site *Site, dt time.Duration, rng *rand.Rand) {
	// Occasionally switch mode, like a real dispatch schedule would.
	if rng.Float64() < 0.02 {
		site.Status = []string{"charging", "discharging", "idle"}[rng.Intn(3)]
	}
	// Turn around before hitting the limits.
	if site.SoC <= 10 && site.Status == "discharging" {
		site.Status = "charging"
	}
	if site.SoC >= 90 && site.Status == "charging" {
		site.Status = "discharging"
	}

	target := 0.0
	switch site.Status {
	case "charging":
		target = -site.MaxPowerMW * (0.6 + 0.4*rng.Float64())
	case "discharging":
		target = site.MaxPowerMW * (0.6 + 0.4*rng.Float64())
	}
	// Move smoothly towards the target power instead of jumping.
	site.PowerMW += (target - site.PowerMW) * 0.3
	site.PowerMW = round(site.PowerMW, 2)

	// Energy = power * time. Discharging lowers SoC, charging raises it.
	energyMWh := site.PowerMW * dt.Hours()
	site.SoC -= energyMWh / site.CapacityMWh * 100
	site.SoC = round(clamp(site.SoC, 0, 100), 2)
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func round(v float64, decimals int) float64 {
	p := 1.0
	for i := 0; i < decimals; i++ {
		p *= 10
	}
	if v < 0 {
		return float64(int64(v*p-0.5)) / p
	}
	return float64(int64(v*p+0.5)) / p
}
