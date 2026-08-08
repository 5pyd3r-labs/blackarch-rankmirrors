// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

const mirrorlistURL = "https://www.blackarch.org/blackarch-mirrorlist"

// FetchMirrorlist downloads the current BlackArch mirrorlist from
// the official source and returns its raw text content.
func FetchMirrorlist() (string, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(mirrorlistURL)
	if err != nil {
		return "", fmt.Errorf("fetching mirrorlist: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d fetching mirrorlist", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("reading mirrorlist body: %w", err)
	}

	return string(body), nil
}
