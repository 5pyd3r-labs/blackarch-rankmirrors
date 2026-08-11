// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"bufio"
	"strings"
)

// Mirror represents a single BlackArch mirror entry.
type Mirror struct {
	Country string
	URL     string
	Active  bool // true if uncommented in the source file
}

// MirrorFile holds the full parsed mirrorlist, grouped by country,
// preserving original ordering for faithful re-rendering.
type MirrorFile struct {
	Countries map[string][]Mirror
	Order     []string // country names in the order they appeared
}

// ParseMirrorlist parses raw BlackArch mirrorlist text (the format
// used by blackarch.org/blackarch-mirrorlist) into a MirrorFile.
func ParseMirrorlist(raw string) MirrorFile {
	mf := MirrorFile{
		Countries: make(map[string][]Mirror),
	}

	scanner := bufio.NewScanner(strings.NewReader(raw))
	currentCountry := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		if line == "" {
			continue
		}

		// Country header: "# CountryName" (not "#Server = ...")
		if strings.HasPrefix(line, "#") && !strings.Contains(line, "Server") {
			country := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			if country == "" {
				continue
			}
			currentCountry = country
			if _, exists := mf.Countries[currentCountry]; !exists {
				mf.Order = append(mf.Order, currentCountry)
				mf.Countries[currentCountry] = []Mirror{}
			}
			continue
		}

		// Mirror line: "Server = ..." or "#Server = ..."
		active := true
		l := line
		if strings.HasPrefix(l, "#") {
			active = false
			l = strings.TrimPrefix(l, "#")
			l = strings.TrimSpace(l)
		}

		if !strings.HasPrefix(l, "Server") {
			continue // unrecognized line, skip
		}

		parts := strings.SplitN(l, "=", 2)
		if len(parts) != 2 {
			continue // malformed, skip per earlier decision (no validation)
		}

		url := strings.TrimSpace(parts[1])
		if url == "" || currentCountry == "" {
			continue
		}

		m := Mirror{
			Country: currentCountry,
			URL:     url,
			Active:  active,
		}
		mf.Countries[currentCountry] = append(mf.Countries[currentCountry], m)
	}

	return mf
}

// AllMirrors flattens MirrorFile into a single slice, useful for
// feeding into the probe/rank stages.
func (mf MirrorFile) AllMirrors() []Mirror {
	var all []Mirror
	for _, country := range mf.Order {
		all = append(all, mf.Countries[country]...)
	}
	return all
}

// FilterExcludedCountries returns mirrors whose Country is not in
// the excluded set (case-insensitive match).
func FilterExcludedCountries(mirrors []Mirror, excluded []string) []Mirror {
	if len(excluded) == 0 {
		return mirrors
	}
	skip := make(map[string]bool, len(excluded))
	for _, c := range excluded {
		skip[strings.ToLower(strings.TrimSpace(c))] = true
	}
	var kept []Mirror
	for _, m := range mirrors {
		if !skip[strings.ToLower(m.Country)] {
			kept = append(kept, m)
		}
	}
	return kept
}

// FilterIncludedCountries returns only mirrors whose Country is in
// the included set (case-insensitive match).
func FilterIncludedCountries(mirrors []Mirror, included []string) []Mirror {
	if len(included) == 0 {
		return mirrors
	}
	keep := make(map[string]bool, len(included))
	for _, c := range included {
		keep[strings.ToLower(strings.TrimSpace(c))] = true
	}
	var kept []Mirror
	for _, m := range mirrors {
		if keep[strings.ToLower(m.Country)] {
			kept = append(kept, m)
		}
	}
	return kept
}
