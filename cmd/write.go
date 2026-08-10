// Copyright (C) 2026 5pyd3r
// Licensed under the GNU General Public License v3.0
// See LICENSE file for full terms.

package rankmirrors

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

const mirrorlistPath = "/etc/pacman.d/blackarch-mirrorlist"
const backupPath = mirrorlistPath + ".bak"

// ConfirmAndWrite prompts the user (unless force is set), backs up
// the existing mirrorlist, and writes the new ranked content.
// Caller must have already verified root privileges.
func ConfirmAndWrite(ranked []ProbeResult, full MirrorFile, force bool) error {
	if !force {
		fmt.Printf("==> Write ranked mirrorlist to %s? [y/N]: ", mirrorlistPath)
		reader := bufio.NewReader(os.Stdin)
		line, _ := reader.ReadString('\n')
		if len(line) == 0 || (line[0] != 'y' && line[0] != 'Y') {
			fmt.Println("[#] Aborted, nothing written.")
			return nil
		}
	}

	if err := backupExisting(); err != nil {
		return fmt.Errorf("backup failed, aborting: %w", err)
	}
	fmt.Printf("[#] Backed up existing file to %s\n", backupPath)

	f, err := os.Create(mirrorlistPath)
	if err != nil {
		return fmt.Errorf("opening %s for write: %w", mirrorlistPath, err)
	}
	defer f.Close()

	if err := RenderMirrorlist(f, ranked, full); err != nil {
		return fmt.Errorf("writing mirrorlist: %w", err)
	}

	fmt.Println("[#] /etc/pacman.d/blackarch-mirrorlist updated successfully.")
	return nil
}

// backupExisting copies the current mirrorlist to a fixed .bak path,
// overwriting any prior backup. Must succeed before the real file
// is touched.
func backupExisting() error {
	src, err := os.Open(mirrorlistPath)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(backupPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
