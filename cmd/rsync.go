// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"bufio"
	"fmt"
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	rsyncMaxProtocol = 32
	rsyncMinProtocol = 27
	rsyncDefaultPort = "873"
)

// rsyncSupportedDigests lists digests this client understands, in
// preference order, used only for protocol >= 32 negotiation.
var rsyncSupportedDigests = []string{
	"sha512",
	"sha256",
	"sha1",
	"md5",
	"md4",
}

// probeRsync performs an rsync daemon handshake against m and
// returns a ProbeResult consistent with probeSingle's http(s)
// results, so the rest of the pipeline (rank, render) doesn't need
// to know or care which protocol produced it.
func probeRsync(m Mirror, timeout time.Duration, ipVersion string) ProbeResult {
	start := time.Now()

	u, err := url.Parse(m.URL)
	if err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	if u.Scheme != "rsync" {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	host := u.Hostname()
	if host == "" {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	module := strings.Trim(u.Path, "/")
	if module == "" {
		return ProbeResult{Mirror: m, Reachable: false}
	}
	if idx := strings.Index(module, "/"); idx != -1 {
		module = module[:idx]
	}

	port := u.Port()
	if port == "" {
		port = rsyncDefaultPort
	}

	network := "tcp"
	if ipVersion == "4" {
		network = "tcp4"
	} else if ipVersion == "6" {
		network = "tcp6"
	}

	conn, err := net.DialTimeout(network, net.JoinHostPort(host, port), timeout)
	if err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}
	defer conn.Close()

	// The overall handshake also needs to respect the timeout, not
	// just the initial dial — set a deadline on the connection.
	if err := conn.SetDeadline(start.Add(timeout)); err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	r := bufio.NewReader(conn)

	greeting, err := r.ReadString('\n')
	if err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}
	greeting = strings.TrimSpace(greeting)

	fields := strings.Fields(greeting)
	if len(fields) < 2 || fields[0] != "@RSYNCD:" {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	versionParts := strings.SplitN(fields[1], ".", 2)
	serverVersion, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	version := serverVersion
	if version > rsyncMaxProtocol {
		version = rsyncMaxProtocol
	}
	if version < rsyncMinProtocol {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	digest := ""
	if version >= 32 {
		digest = rsyncNegotiateDigest(fields[2:])
		if digest == "" {
			// No mutually supported digest; treat as unreachable
			// rather than erroring the whole run.
			return ProbeResult{Mirror: m, Reachable: false}
		}
	}

	if version >= 32 {
		_, err = fmt.Fprintf(conn, "@RSYNCD: %d.0 %s\n", version, digest)
	} else {
		_, err = fmt.Fprintf(conn, "@RSYNCD: %d.0\n", version)
	}
	if err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	if _, err := fmt.Fprintf(conn, "%s\n", module); err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	response, err := r.ReadString('\n')
	if err != nil {
		return ProbeResult{Mirror: m, Reachable: false}
	}
	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "@ERROR") {
		return ProbeResult{Mirror: m, Reachable: false}
	}

	return ProbeResult{Mirror: m, Reachable: true, Latency: time.Since(start)}
}

func rsyncNegotiateDigest(server []string) string {
	serverSet := make(map[string]struct{}, len(server))
	for _, d := range server {
		serverSet[strings.ToLower(d)] = struct{}{}
	}
	for _, d := range rsyncSupportedDigests {
		if _, ok := serverSet[d]; ok {
			return d
		}
	}
	return ""
}
