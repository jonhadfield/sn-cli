package main

import (
	"fmt"

	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	sncli "github.com/jonhadfield/sn-cli/internal/sncli"
	"github.com/urfave/cli/v2"
)

// maxDuplicateGroupsListed caps how many sets of duplicates are listed before
// the rest are summarised, so a large account does not fill the terminal.
const maxDuplicateGroupsListed = 20

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

	writeDuplicatesKept(c, found.KeptNoOriginal)

	toDelete := found.Deleted()

	if len(toDelete) == 0 {
		_, _ = fmt.Fprintln(c.App.Writer, "no duplicate notes found")

		return nil
	}

	writeDuplicateGroups(c, found.Groups, len(toDelete), dryRun)

	if dryRun {
		return nil
	}

	if !c.Bool("yes") && !confirmDeleteDuplicates(c, len(toDelete)) {
		return nil
	}

	del := sncli.DeleteDuplicateNotesConfig{Session: &sess, Debug: opts.debug}

	out, err := del.Run()
	if err != nil {
		return fmt.Errorf("failed to delete duplicate notes: %w", err)
	}

	_, _ = fmt.Fprintln(c.App.Writer, color.Green.Sprintf("%d duplicate notes deleted", len(out.Deleted())))

	return nil
}

func writeDuplicateGroups(c *cli.Context, groups []sncli.DuplicateGroup, numToDelete int, dryRun bool) {
	heading := fmt.Sprintf("%d notes to delete, in %d sets of duplicates:", numToDelete, len(groups))
	if dryRun {
		heading = fmt.Sprintf("%d notes would be deleted, in %d sets of duplicates:", numToDelete, len(groups))
	}

	_, _ = fmt.Fprintln(c.App.Writer, heading)

	for x, group := range groups {
		if x == maxDuplicateGroupsListed {
			_, _ = fmt.Fprintln(c.App.Writer, color.Gray.Sprintf("  ... and %d more sets", len(groups)-maxDuplicateGroupsListed))

			break
		}

		_, _ = fmt.Fprintf(c.App.Writer, "  keep    %s  %s  %s\n",
			group.Keep.UpdatedDate(), group.Keep.UUID, group.Keep.Title)

		for _, dup := range group.Delete {
			_, _ = fmt.Fprintf(c.App.Writer, "  %s  %s  %s  %s\n",
				color.Red.Sprint("delete"), dup.UpdatedDate(), dup.UUID, dup.Title)
		}
	}
}

func writeDuplicatesKept(c *cli.Context, kept []sncli.NoteSummary) {
	if len(kept) == 0 {
		return
	}

	_, _ = fmt.Fprintln(c.App.Writer, color.Yellow.Sprintf(
		"%d copies kept, as the note each was made from is no longer in the account:", len(kept)))

	for x, dup := range kept {
		if x == maxDuplicateGroupsListed {
			_, _ = fmt.Fprintln(c.App.Writer, color.Gray.Sprintf("  ... and %d more", len(kept)-maxDuplicateGroupsListed))

			break
		}

		_, _ = fmt.Fprintf(c.App.Writer, "  %s  %s\n", dup.UUID, dup.Title)
	}
}

func confirmDeleteDuplicates(c *cli.Context, num int) bool {
	_, _ = fmt.Fprintf(c.App.Writer, "delete %d notes? ", num)

	var input string

	if _, err := fmt.Scanln(&input); err != nil {
		return false
	}

	return sncli.StringInSlice(input, []string{"y", "yes"}, false)
}
