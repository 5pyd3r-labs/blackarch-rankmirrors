// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package main

import (
	"flag"
	"fmt"
	"os"

	rankmirrors "github.com/5pyd3r-labs/blackarch-rankmirrors/cmd"
)

func main() {
	allowHTTP := flag.Bool("allow-http", false, "also test http:// mirrors, not just https://")
	n := flag.Int("n", 5, "number of top mirrors to output/write")
	update := flag.Bool("update", false, "write ranked mirrorlist to /etc/pacman.d/blackarch-mirrorlist (requires root)")
	force := flag.Bool("force", false, "skip confirmation prompt (only meaningful with --update)")
	flag.Parse()

	if *update && os.Geteuid() != 0 {
		fmt.Fprintln(os.Stderr, "error: --update requires root privileges, rerun with: sudo barch-rankmirrors --update")
		os.Exit(1)
	}

	fmt.Println("Fetching mirrorlist from blackarch.org...")

	data, err := rankmirrors.FetchMirrorlist()
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	mf := rankmirrors.ParseMirrorlist(data)
	all := mf.AllMirrors()

	mode := "https-only"
	if *allowHTTP {
		mode = "http+https"
	}
	fmt.Printf("Parsed %d mirrors across %d countries\n", len(all), len(mf.Order))
	fmt.Printf("Probing mirrors (%s)...\n", mode)

	results := rankmirrors.ProbeMirrors(all, *allowHTTP)

	reachableCount := rankmirrors.ReachableCount(results)
	if reachableCount == 0 {
		fmt.Fprintln(os.Stderr, "error: no reachable mirrors found. Try --allow-http, or check your connection.")
		os.Exit(1)
	}

	ranked := rankmirrors.RankMirrors(results, *n)

	fmt.Printf("\n%d of %d tested mirrors reachable, showing top %d:\n\n", reachableCount, len(results), len(ranked))
	fmt.Printf("%-6s %-10s %s\n", "RANK", "TIME", "MIRROR")
	for i, r := range ranked {
		fmt.Printf("%-6d %-10s %s\n", i+1, r.Latency.Round(1000000), r.Mirror.URL)
	}

	if !*update {
		fmt.Println("\nRun with --update to write to /etc/pacman.d/blackarch-mirrorlist.")
		return
	}

	if err := rankmirrors.ConfirmAndWrite(ranked, mf, *force); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
