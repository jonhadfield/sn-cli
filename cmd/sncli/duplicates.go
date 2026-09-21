package main

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/urfave/cli/v2"
)

// maxDuplicatesListed caps how many duplicates are listed before the rest are
// summarised, so a large account does not fill the terminal.
const maxDuplicatesListed = 20

func processDeleteDuplicates(c *cli.Context, opts configOptsOutput) error {
	sess, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	if sess.CacheDBPath, err = cache.GenCacheDBPath(sess, opts.cacheDBDir, snAppName); err != nil {
		return err
	}

	dryRun := c.Bool("dry-run")

	// find the duplicates first, so they can be shown before anything is
	// deleted
	find := sncli.DeleteDuplicateNotesConfig{Session: &sess, DryRun: true, Debug: opts.debug}

	found, err := find.Run()
	if err != nil {
		return fmt.Errorf("failed to find duplicate notes: %w", err)
	}

	writeDuplicatesKept(c, found.KeptOriginalMissing)

	if len(found.Deleted) == 0 {
		_, _ = fmt.Fprintln(c.App.Writer, "no duplicate notes found")

		return nil
	}

	writeDuplicatesFound(c, found.Deleted, dryRun)

	if dryRun {
		return nil
	}

	if !c.Bool("yes") && !confirmDeleteDuplicates(c, len(found.Deleted)) {
		return nil
	}

	del := sncli.DeleteDuplicateNotesConfig{Session: &sess, Debug: opts.debug}

	out, err := del.Run()
	if err != nil {
		return fmt.Errorf("failed to delete duplicate notes: %w", err)
	}

	_, _ = fmt.Fprintln(c.App.Writer, color.Green.Sprintf("%d duplicate notes deleted", len(out.Deleted)))

	return nil
}

func writeDuplicatesFound(c *cli.Context, dups []sncli.DuplicateNote, dryRun bool) {
	heading := fmt.Sprintf("%d duplicate notes to delete:", len(dups))
	if dryRun {
		heading = fmt.Sprintf("%d duplicate notes would be deleted:", len(dups))
	}

	_, _ = fmt.Fprintln(c.App.Writer, heading)

	for x, dup := range dups {
		if x == maxDuplicatesListed {
			_, _ = fmt.Fprintln(c.App.Writer, color.Gray.Sprintf("  ... and %d more", len(dups)-maxDuplicatesListed))

			break
		}

		_, _ = fmt.Fprintf(c.App.Writer, "  %s  %s\n", dup.UUID, dup.Title)
	}
}

func writeDuplicatesKept(c *cli.Context, kept []sncli.DuplicateNote) {
	if len(kept) == 0 {
		return
	}

	_, _ = fmt.Fprintln(c.App.Writer, color.Yellow.Sprintf(
		"%d duplicates kept, as the note each was copied from is no longer in the account:", len(kept)))

	for x, dup := range kept {
		if x == maxDuplicatesListed {
			_, _ = fmt.Fprintln(c.App.Writer, color.Gray.Sprintf("  ... and %d more", len(kept)-maxDuplicatesListed))

			break
		}

		_, _ = fmt.Fprintf(c.App.Writer, "  %s  %s\n", dup.UUID, dup.Title)
	}
}

func confirmDeleteDuplicates(c *cli.Context, num int) bool {
	_, _ = fmt.Fprintf(c.App.Writer, "delete %d duplicate notes? ", num)

	var input string

	if _, err := fmt.Scanln(&input); err != nil {
		return false
	}

	return sncli.StringInSlice(input, []string{"y", "yes"}, false)
}
