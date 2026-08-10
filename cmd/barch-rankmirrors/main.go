// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	rankmirrors "github.com/5pyd3r-labs/blackarch-rankmirrors/cmd"
)

func main() {
	allowHTTP := flag.Bool("allow-http", false, "also test http:// mirrors, not just https://")
	n := flag.Int("n", 5, "number of top mirrors to output/write")
	update := flag.Bool("update", false, "write ranked mirrorlist to /etc/pacman.d/blackarch-mirrorlist (requires root)")
	force := flag.Bool("force", false, "skip confirmation prompt (only meaningful with --update)")
	excludeCountry := flag.String("exclude-country", "", "comma-separated list of countries to exclude (e.g. 'China,Russia')")
	timeout := flag.Duration("per-mirror-timeout", 5*time.Second, "per-mirror probe timeout (e.g. 5s, 500ms)")
	rankAll := flag.Bool("rank-all", false, "show all reachable mirrors ranked, ignoring --n (display only, --update still uses --n)")
	output := flag.String("output", "", "save the full run output (everything printed to screen) to this file")
	flag.Parse()

	start := time.Now()

	var out io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[X] error creating output file:", err)
			os.Exit(1)
		}
		defer f.Close()
		out = io.MultiWriter(os.Stdout, f)
	}

	if *update && os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "[X] error: --update requires root privileges, rerun with: sudo barch-rankmirrors --update")
		os.Exit(1)
	}

	fmt.Fprintln(out)
	fmt.Fprintf(out, "[#] Command: barch-rankmirrors %s\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintln(out, "[#] Fetching mirrorlist from blackarch.org...")

	data, err := rankmirrors.FetchMirrorlist()
	if err != nil {
		fmt.Fprintln(os.Stderr, "[X] error:", err)
		os.Exit(1)
	}

	mf := rankmirrors.ParseMirrorlist(data)
	all := mf.AllMirrors()

	if *excludeCountry != "" {
		excluded := strings.Split(*excludeCountry, ",")
		all = rankmirrors.FilterExcludedCountries(all, excluded)
		fmt.Fprintf(out, "[-] Excluding countries: %s\n", *excludeCountry)
	}

	mode := "https-only"
	if *allowHTTP {
		mode = "http+https"
	}
	fmt.Fprintf(out, "[+] Parsed %d mirrors across %d countries\n", len(all), len(mf.Order))
	fmt.Fprintf(out, "[#] Probing mirrors (%s)...\n", mode)

	results := rankmirrors.ProbeMirrors(all, *allowHTTP, *timeout)

	reachableCount := rankmirrors.ReachableCount(results)
	if reachableCount == 0 {
		fmt.Fprintln(os.Stderr, "[X] error: no reachable mirrors found. Try --allow-http, or check your connection.")
		os.Exit(1)
	}

	ranked := rankmirrors.RankMirrors(results, *n)

	displayed := ranked
	if *rankAll {
		displayed = rankmirrors.RankMirrors(results, reachableCount)
	}

	fmt.Fprintf(out, "\n==> %d of %d tested mirrors reachable, showing top %d:\n\n", reachableCount, len(results), len(displayed))
	fmt.Fprintf(out, "%-6s %-10s %-15s %s\n", "RANK", "TIME", "COUNTRY", "MIRROR")
	for i, r := range displayed {
		fmt.Fprintf(out, "%-6d %-10s %-15s %s\n", i+1, r.Latency.Round(time.Millisecond), r.Mirror.Country, r.Mirror.URL)
	}

	elapsed := time.Since(start)
	fmt.Fprintf(out, "\n[#] Completed in %s\n", elapsed.Round(time.Millisecond))

	if *output != "" {
		fmt.Printf("[#] Output written to %s\n", *output)
	}

	if !*update {
		fmt.Fprintln(out, "[*] Run with --update to write to /etc/pacman.d/blackarch-mirrorlist.")
		return
	}

	if err := rankmirrors.ConfirmAndWrite(ranked, mf, *force); err != nil {
		fmt.Fprintln(os.Stderr, "[X] error:", err)
		os.Exit(1)
	}
}
