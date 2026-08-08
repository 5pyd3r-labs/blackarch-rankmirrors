// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import "sort"

// RankMirrors filters to reachable results, sorts by latency
// ascending, and returns at most n entries (clamped to however
// many are actually reachable).
func RankMirrors(results []ProbeResult, n int) []ProbeResult {
	var reachable []ProbeResult
	for _, r := range results {
		if r.Reachable {
			reachable = append(reachable, r)
		}
	}

	sort.Slice(reachable, func(i, j int) bool {
		return reachable[i].Latency < reachable[j].Latency
	})

	if n > len(reachable) {
		n = len(reachable)
	}

	return reachable[:n]
}

// ReachableCount returns how many results were reachable, before
// ranking/clamping.
func ReachableCount(results []ProbeResult) int {
	count := 0
	for _, r := range results {
		if r.Reachable {
			count++
		}
	}
	return count
}
