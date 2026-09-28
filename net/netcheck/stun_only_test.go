// Copyright (c) Tailscale Inc & contributors
// SPDX-License-Identifier: BSD-3-Clause

package netcheck

import (
	"testing"
	"time"

	"tailscale.com/net/netmon"
	"tailscale.com/tailcfg"
)

func TestSTUNOnlyPreferredDERP(t *testing.T) {
	dm := &tailcfg.DERPMap{Regions: map[tailcfg.DERPRegionID]*tailcfg.DERPRegion{
		1: {RegionID: 1, Nodes: []*tailcfg.DERPNode{{Name: "one", RegionID: 1}}},
		2: {RegionID: 2, Nodes: []*tailcfg.DERPNode{{Name: "two", RegionID: 2}}},
		3: {RegionID: 3, Nodes: []*tailcfg.DERPNode{{Name: "stun", RegionID: 3, STUNOnly: true}}},
	}}
	plan := makeProbePlan(dm, &netmon.State{HaveV4: true}, nil, 0)
	found := false
	for _, probes := range plan {
		for _, p := range probes {
			found = found || p.node == "stun"
		}
	}
	if !found {
		t.Fatal("STUN-only region omitted from discovery probes")
	}
	for _, tt := range []struct {
		name         string
		prev, forced tailcfg.DERPRegionID
		latencies    map[tailcfg.DERPRegionID]time.Duration
		want         tailcfg.DERPRegionID
	}{
		{"fast STUN", 0, 0, map[tailcfg.DERPRegionID]time.Duration{1: 100 * time.Millisecond, 2: 50 * time.Millisecond, 3: time.Millisecond}, 2},
		{"stale STUN home", 3, 0, map[tailcfg.DERPRegionID]time.Duration{1: 100 * time.Millisecond, 2: 50 * time.Millisecond, 3: time.Millisecond}, 2},
		{"forced STUN", 0, 3, map[tailcfg.DERPRegionID]time.Duration{1: 100 * time.Millisecond, 2: 50 * time.Millisecond, 3: time.Millisecond}, 2},
		{"only STUN reachable", 3, 3, map[tailcfg.DERPRegionID]time.Duration{3: time.Millisecond}, 0},
		{"nothing reachable stale home", 3, 3, nil, 0},
		{"all STUN map", 3, 3, map[tailcfg.DERPRegionID]time.Duration{3: time.Millisecond}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			probeMap := dm
			if tt.name == "all STUN map" {
				probeMap = dm.Clone()
				delete(probeMap.Regions, 1)
				delete(probeMap.Regions, 2)
			}
			now := time.Now()
			c := &Client{ForcePreferredDERP: tt.forced, last: &Report{PreferredDERP: tt.prev}}
			r := &Report{RegionLatency: tt.latencies}
			rs := &reportState{start: now.Add(-time.Second), opts: &GetReportOpts{GetLastDERPActivity: func(tailcfg.DERPRegionID) time.Time { return now }}}
			c.addReportHistoryAndSetPreferredDERP(rs, r, probeMap.View(), now)
			if r.PreferredDERP != tt.want {
				t.Fatalf("home=%d, want %d", r.PreferredDERP, tt.want)
			}
		})
	}
}
