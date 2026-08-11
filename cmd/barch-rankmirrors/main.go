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
	output := flag.String("output", "", "save whatever is printed to stdout to this file as well")
	mirrorsOnly := flag.Bool("mirrors-only", false, "print only the rendered mirrorlist to stdout (no logs/table), safe to pipe into a file")
	flag.Parse()

	if *mirrorsOnly && *update {
		fmt.Fprintln(os.Stderr, "[X] error: --mirrors-only and --update are mutually exclusive — use --mirrors-only to pipe into your own write command, or --update to let the tool write directly.")
		os.Exit(1)
	}

	start := time.Now()

	var stdoutWriter io.Writer = os.Stdout
	if *output != "" {
		f, err := os.Create(*output)
		if err != nil {
			fmt.Fprintln(os.Stderr, "[X] error creating output file:", err)
			os.Exit(1)
		}
		defer f.Close()
		stdoutWriter = io.MultiWriter(os.Stdout, f)
	}

	var logOut io.Writer = stdoutWriter
	if *mirrorsOnly {
		logOut = os.Stderr
	}

	if *update && os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "[X] error: --update requires root privileges, rerun with: sudo barch-rankmirrors --update")
		os.Exit(1)
	}

	fmt.Fprintln(logOut)
	fmt.Fprintf(logOut, "[#] Command: barch-rankmirrors %s\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintln(logOut, "[#] Fetching mirrorlist from blackarch.org...")

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
		fmt.Fprintf(logOut, "[-] Excluding countries: %s\n", *excludeCountry)
	}

	mode := "https-only"
	if *allowHTTP {
		mode = "http+https"
	}
	fmt.Fprintf(logOut, "[+] Parsed %d mirrors across %d countries\n", len(all), len(mf.Order))
	fmt.Fprintf(logOut, "[#] Probing mirrors (%s)...\n", mode)

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

	fmt.Fprintf(logOut, "\n==> %d of %d tested mirrors reachable, showing top %d:\n\n", reachableCount, len(results), len(displayed))
	fmt.Fprintf(logOut, "%-6s %-10s %-15s %s\n", "RANK", "TIME", "COUNTRY", "MIRROR")
	for i, r := range displayed {
		fmt.Fprintf(logOut, "%-6d %-10s %-15s %s\n", i+1, r.Latency.Round(time.Millisecond), r.Mirror.Country, r.Mirror.URL)
	}

	elapsed := time.Since(start)
	fmt.Fprintf(logOut, "\n[#] Completed in %s\n", elapsed.Round(time.Millisecond))

	if *mirrorsOnly {
		if err := rankmirrors.RenderMirrorlist(stdoutWriter, ranked, mf); err != nil {
			fmt.Fprintln(os.Stderr, "[X] error rendering mirrorlist:", err)
			os.Exit(1)
		}
	}

	if !*update {
		fmt.Fprintln(logOut, "[*] Run with --update to write to /etc/pacman.d/blackarch-mirrorlist.")
		return
	}

	if err := rankmirrors.ConfirmAndWrite(ranked, mf, *force); err != nil {
		fmt.Fprintln(os.Stderr, "[X] error:", err)
		os.Exit(1)
	}
}
