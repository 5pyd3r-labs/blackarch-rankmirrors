// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"context"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

const probeConcurrency = 20

// ProbeResult holds the outcome of testing a single mirror.
type ProbeResult struct {
	Mirror    Mirror
	Reachable bool
	Latency   time.Duration
}

// ProbeMirrors tests all given mirrors concurrently for reachability
// and response time. If allowHTTP is false, only https:// mirrors
// are tested; http:// is included when allowHTTP is true. rsync://
// mirrors are only tested when allowRsync is true, and are probed
// via a raw daemon handshake (see rsync.go) rather than the shared
// http.Client. Any other scheme is skipped entirely. timeout
// controls the per-mirror request timeout.
func ProbeMirrors(mirrors []Mirror, allowHTTP bool, allowRsync bool, timeout time.Duration, ipVersion string) []ProbeResult {
	var candidates []Mirror
	for _, m := range mirrors {
		if isTestableURL(m.URL, allowHTTP, allowRsync) {
			candidates = append(candidates, m)
		}
	}

	results := make([]ProbeResult, len(candidates))
	sem := make(chan struct{}, probeConcurrency)
	var wg sync.WaitGroup

	client := newClient(timeout, ipVersion)

	for i, m := range candidates {
		wg.Add(1)
		sem <- struct{}{}
		go func(idx int, mirror Mirror) {
			defer wg.Done()
			defer func() { <-sem }()

			if strings.HasPrefix(mirror.URL, "rsync://") {
				results[idx] = probeRsync(mirror, timeout, ipVersion)
			} else {
				results[idx] = probeSingle(client, mirror)
			}
		}(i, m)
	}

	wg.Wait()
	return results
}

// isTestableURL filters out unsupported protocols. https:// is
// always testable; http:// and rsync:// are gated behind their
// respective allow flags.
func isTestableURL(url string, allowHTTP bool, allowRsync bool) bool {
	switch {
	case strings.HasPrefix(url, "https://"):
		return true
	case strings.HasPrefix(url, "http://"):
		return allowHTTP
	case strings.HasPrefix(url, "rsync://"):
		return allowRsync
	default:
		return false // ftp://, ftpes://, gopher://, etc.
	}
}

// probeSingle performs a single HEAD request and measures latency.
func probeSingle(client *http.Client, m Mirror) ProbeResult {
	dbURL := deriveDBURL(m.URL)

	start := time.Now()
	resp, err := client.Head(dbURL)
	elapsed := time.Since(start)

	if err != nil || resp.StatusCode != http.StatusOK {
		return ProbeResult{Mirror: m, Reachable: false}
	}
	defer resp.Body.Close()

	return ProbeResult{Mirror: m, Reachable: true, Latency: elapsed}
}

// deriveDBURL turns a mirrorlist template URL (with $repo/$arch
// placeholders) into an actual fetchable path for probing, using
// the "blackarch" repo and x86_64 arch, pointing at the db file.
func deriveDBURL(templateURL string) string {
	url := strings.ReplaceAll(templateURL, "$repo", "blackarch")
	url = strings.ReplaceAll(url, "$arch", "x86_64")
	if !strings.HasSuffix(url, "/") {
		url += "/"
	}
	return url + "blackarch.db"
}

func newClient(timeout time.Duration, ipVersion string) *http.Client {
	dialer := &net.Dialer{}

	transport := &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			forced := network
			switch ipVersion {
			case "4":
				forced = "tcp4"
			case "6":
				forced = "tcp6"
			}
			return dialer.DialContext(ctx, forced, addr)
		},
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: transport,
	}
}
