// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"sort"
	"time"
)

// phase2SubsetMultiplier controls how many phase-1 candidates get
// carried into phase 2, relative to n. A larger multiplier gives
// the median-refinement step more mirrors to choose from, at the
// cost of more re-probing.
const phase2SubsetMultiplier = 3

// RankMirrorsRefined performs the full two-phase ranking: an initial
// single-probe pass (already reflected in results) is used to pick
// a subset of the most promising mirrors, that subset is re-probed
// several times via RefineWithMedian to get a more reliable latency
// figure, and the final ranking is computed from those refined
// results. This trades a bit of extra time for a ranking that is
// less sensitive to a single noisy probe.
func RankMirrorsRefined(results []ProbeResult, n int, timeout time.Duration, ipVersion string) []ProbeResult {
	reachableCount := ReachableCount(results)

	subsetSize := n * phase2SubsetMultiplier
	if subsetSize > reachableCount {
		subsetSize = reachableCount
	}

	subset := RankMirrors(results, subsetSize)
	refined := RefineWithMedian(subset, timeout, ipVersion)

	return RankMirrors(refined, n)
}

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

// rankAllRefineCap limits how many reachable mirrors get refined
// when --rank-all is set, so a broad, unfiltered run doesn't trigger
// refinement on dozens of mirrors nobody will realistically pick.
// Mirrors beyond the cap keep their phase-1 single-probe latency.
const rankAllRefineCap = 30

// RankAllRefined returns every reachable mirror, sorted by latency,
// with the fastest rankAllRefineCap of them refined via
// RefineWithMedian for consistency with RankMirrorsRefined's output.
// Mirrors beyond the cap are included using their phase-1 latency.
func RankAllRefined(results []ProbeResult, timeout time.Duration, ipVersion string) []ProbeResult {
	reachableCount := ReachableCount(results)

	capSize := rankAllRefineCap
	if capSize > reachableCount {
		capSize = reachableCount
	}

	toRefine := RankMirrors(results, capSize)
	refined := RefineWithMedian(toRefine, timeout, ipVersion)

	rest := RankMirrors(results, reachableCount)[capSize:]

	combined := append(refined, rest...)

	sort.Slice(combined, func(i, j int) bool {
		return combined[i].Latency < combined[j].Latency
	})

	return combined
}
