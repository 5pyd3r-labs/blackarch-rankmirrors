// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

const banner = `######################################################
###                                                ###
##         BlackArch Repository Mirrorlist          ##
###                                                ###
######################################################`

// RenderMirrorlist writes the full pacman.d/blackarch-mirrorlist
// content: a curated active block from the ranked results (each
// mirror preceded by a country comment line), followed by the
// complete original mirrorlist (re-commented, structure and order
// preserved) for reference.
func RenderMirrorlist(w io.Writer, ranked []ProbeResult, full MirrorFile) error {
	bw := bufio.NewWriter(w)

	fmt.Fprintln(bw, "#---------------------------------------------------")
	fmt.Fprintf(bw, "# Current mirrors in use - updated on %s\n",
		time.Now().Format("2006-01-02"))
	fmt.Fprintln(bw, "#---------------------------------------------------")
	fmt.Fprintln(bw)

	for _, r := range ranked {
		fmt.Fprintf(bw, "# %s\n", strings.ToLower(r.Mirror.Country))
		fmt.Fprintf(bw, "Server = %s\n", r.Mirror.URL)
	}
	fmt.Fprintln(bw)

	fmt.Fprintln(bw, banner)
	fmt.Fprintln(bw)

	for _, country := range full.Order {
		fmt.Fprintf(bw, "# %s\n", country)
		for _, m := range full.Countries[country] {
			fmt.Fprintf(bw, "#Server = %s\n", m.URL)
		}
		fmt.Fprintln(bw)
	}

	return bw.Flush()
}
