package main

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"github.com/asdine/storm/v3"
	"github.com/gookit/color"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/jonhadfield/gosn-v2/common"
	"github.com/jonhadfield/gosn-v2/items"
	"github.com/jonhadfield/gosn-v2/session"
	"github.com/pterm/pterm"
)

// TagStats holds statistics about a tag
type TagStats struct {
	Title     string
	UUID      string
	NoteCount int
	CreatedAt string
}

// getItemsViaSync uses cache.Sync to properly load items, handling network errors gracefully
func getItemsViaSync(session *cache.Session, debug bool) (items.Items, items.Items, error) {
	// Sync to load items from cache (and server if available)
	si := cache.SyncInput{
		Session: session,
		Close:   false,
	}

	so, syncErr := cache.Sync(si)

	// Check if we got a database connection even if sync failed
	if so.DB == nil {
		if syncErr != nil {
			return nil, nil, fmt.Errorf("failed to open cache database: %w", syncErr)
		}
		return nil, nil, fmt.Errorf("no cache database available")
	}
	defer so.DB.Close()

	// If sync failed but we have DB, just warn and continue with cache
	if syncErr != nil {
		if debug {
			pterm.Warning.Printf("Sync failed, using cached data only: %v\n", syncErr)
		}
	}

	// Get all items from database
	var allPersistedItems cache.Items
	if err := so.DB.All(&allPersistedItems); err != nil {
		return nil, nil, fmt.Errorf("failed to read from cache: %w", err)
	}

	// Convert to items (session now has keys from sync)
	allItems, err := allPersistedItems.ToItems(session)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to decrypt items: %w", err)
	}

	// Separate tags and notes
	var tags items.Items
	var notes items.Items

	for _, item := range allItems {
		if item.IsDeleted() {
			continue
		}
		switch item.GetContentType() {
		case common.SNItemTypeTag:
			tags = append(tags, item)
		case common.SNItemTypeNote:
			notes = append(notes, item)
		}
	}

	return tags, notes, nil
}

// getItemsFromCache reads tags and notes directly from cache without syncing
func getItemsFromCache(session *cache.Session, debug bool) (items.Items, items.Items, error) {
	// Open cache database directly
	cacheDB, err := storm.Open(session.CacheDBPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open cache database: %w (have you run a sync recently?)", err)
	}
	defer cacheDB.Close()

	// Get all cached items
	var allPersistedItems cache.Items
	if err = cacheDB.All(&allPersistedItems); err != nil {
		return nil, nil, fmt.Errorf("failed to read cached items: %w", err)
	}

	if len(allPersistedItems) == 0 {
		// No cached data - need to sync first
		return nil, nil, fmt.Errorf("no cached data found - please run 'sncli get note' first to populate cache")
	}

	// Load items keys from cache (mimics cache.Sync behavior)
	cachedKeys, err := retrieveItemsKeysFromCache(session.Session, allPersistedItems)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve items keys from cache: %w", err)
	}

	if err = processCachedItemsKeys(session, cachedKeys); err != nil {
		return nil, nil, fmt.Errorf("failed to process cached items keys: %w", err)
	}

	// Convert to items (session now has keys)
	allItems, err := allPersistedItems.ToItems(session)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to convert cached items: %w", err)
	}

	// Separate tags and notes
	var tags items.Items
	var notes items.Items

	if debug {
		pterm.Debug.Printf("Total items in cache: %d\n", len(allItems))
	}

	for _, item := range allItems {
		if item.IsDeleted() {
			continue
		}

		contentType := item.GetContentType()
		if debug {
			pterm.Debug.Printf("Item type: %s, UUID: %s\n", contentType, item.GetUUID())
		}

		switch contentType {
		case common.SNItemTypeTag:
			tags = append(tags, item)
		case common.SNItemTypeNote:
			notes = append(notes, item)
		}
	}

	if debug {
		pterm.Debug.Printf("Found %d tags and %d notes\n", len(tags), len(notes))
	}

	return tags, notes, nil
}

// retrieveItemsKeysFromCache gets items keys from cached items (from gosn-v2/cache)
func retrieveItemsKeysFromCache(s *session.Session, cachedItems cache.Items) (items.EncryptedItems, error) {
	var itemsKeys items.EncryptedItems

	for _, ci := range cachedItems {
		if ci.ContentType == common.SNItemTypeItemsKey && !ci.Deleted {
			itemsKeys = append(itemsKeys, items.EncryptedItem{
				UUID:        ci.UUID,
				Content:     ci.Content,
				ContentType: ci.ContentType,
				ItemsKeyID:  ci.ItemsKeyID,
				EncItemKey:  ci.EncItemKey,
			})
		}
	}

	return itemsKeys, nil
}

// processCachedItemsKeys processes items keys and adds to session (from gosn-v2/cache)
func processCachedItemsKeys(cs *cache.Session, eiks items.EncryptedItems) error {
	if len(eiks) == 0 {
		return nil
	}

	// Decrypt and parse items keys
	iks, err := items.DecryptAndParseItemKeys(cs.MasterKey, eiks)
	if err != nil {
		return err
	}

	// Convert to session items keys
	var syncedItemsKeys []session.SessionItemsKey
	for x := range iks {
		syncedItemsKeys = append(syncedItemsKeys, session.SessionItemsKey{
			UUID:               iks[x].UUID,
			ItemsKey:           iks[x].ItemsKey,
			Default:            iks[x].Default,
			UpdatedAtTimestamp: iks[x].UpdatedAtTimestamp,
			CreatedAtTimestamp: iks[x].CreatedAtTimestamp,
		})
	}

	// Merge with existing items keys in session
	cs.Session.ItemsKeys = mergeItemsKeysSlices(cs.Session.ItemsKeys, syncedItemsKeys)

	// Set default items key using Standard Notes priority logic:
	// 1. Prioritize keys marked as default
	// 2. Fall back to most recent by timestamp
	var defaultItemsKey session.SessionItemsKey
	var latestItemsKey session.SessionItemsKey

	for x := range cs.Session.ItemsKeys {
		key := cs.Session.ItemsKeys[x]

		// Track the most recent key regardless
		if key.CreatedAtTimestamp > latestItemsKey.CreatedAtTimestamp {
			latestItemsKey = key
		}

		// Prefer keys marked as default
		if key.Default {
			defaultItemsKey = key
			break
		}
	}

	// Use default key if found, otherwise use most recent
	if defaultItemsKey.UUID != "" {
		cs.Session.DefaultItemsKey = defaultItemsKey
	} else if latestItemsKey.UUID != "" {
		cs.Session.DefaultItemsKey = latestItemsKey
	}

	return nil
}

// mergeItemsKeysSlices merges two slices of items keys (from gosn-v2/cache)
func mergeItemsKeysSlices(existing, new []session.SessionItemsKey) []session.SessionItemsKey {
	// Create map of existing keys by UUID
	existingMap := make(map[string]session.SessionItemsKey)
	for _, k := range existing {
		existingMap[k.UUID] = k
	}

	// Add or update with new keys
	for _, k := range new {
		existingMap[k.UUID] = k
	}

	// Convert back to slice
	var result []session.SessionItemsKey
	for _, k := range existingMap {
		result = append(result, k)
	}

	return result
}

// getTagUUIDs returns a slice of tag UUIDs for debugging
func getTagUUIDs(tagStats map[string]*TagStats) []string {
	var uuids []string
	for uuid := range tagStats {
		uuids = append(uuids, uuid)
		if len(uuids) >= 10 {
			break
		}
	}
	return uuids
}

// ShowTagCloud displays tags as a visual cloud
func ShowTagCloud(opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Use existing sync-based approach that works for other commands
	// This ensures all processing steps are applied correctly
	rawTags, rawNotes, err := getItemsViaSync(&session, opts.debug)
	if err != nil {
		return err
	}

	// Check first 10 notes for references (always show)
	fmt.Printf("Checking first 10 notes for references...\n")
	foundRefs := false
	if len(rawNotes) > 0 {
		for i := 0; i < 10 && i < len(rawNotes); i++ {
			sampleNote := rawNotes[i].(*items.Note)
			refs := sampleNote.Content.References()
			if len(refs) > 0 {
				foundRefs = true
				fmt.Printf("  Note %d has %d references:\n", i, len(refs))
				for j, ref := range refs {
					if j < 3 {
						shortUUID := ref.UUID
						if len(shortUUID) > 12 {
							shortUUID = shortUUID[:12] + "..."
						}
						fmt.Printf("    Ref %d: Type='%s', UUID='%s'\n", j, ref.ContentType, shortUUID)
					}
				}
				break
			}
		}
	}
	if !foundRefs {
		fmt.Printf("  No references found in first 10 notes\n")
	}

	// CRITICAL: Check if TAGS have references to NOTES (reverse direction)
	fmt.Printf("\nChecking first 10 tags for references to notes...\n")
	foundTagRefs := false
	if len(rawTags) > 0 {
		for i := 0; i < 10 && i < len(rawTags); i++ {
			sampleTag := rawTags[i].(*items.Tag)
			refs := sampleTag.Content.References()
			if len(refs) > 0 {
				foundTagRefs = true
				fmt.Printf("  Tag '%s' has %d references:\n", sampleTag.Content.GetTitle(), len(refs))
				for j, ref := range refs {
					if j < 3 {
						shortUUID := ref.UUID
						if len(shortUUID) > 12 {
							shortUUID = shortUUID[:12] + "..."
						}
						fmt.Printf("    Ref %d: Type='%s', UUID='%s'\n", j, ref.ContentType, shortUUID)
					}
				}
				if !foundTagRefs {
					break
				}
			}
		}
	}
	if !foundTagRefs {
		fmt.Printf("  No references found in first 10 tags\n")
	}

	// Build tag statistics and count from tags (Tag → Note references)
	tagStats := make(map[string]*TagStats)
	noteUUIDs := make(map[string]bool)

	// Build note UUID map for validation
	for _, item := range rawNotes {
		note := item.(*items.Note)
		noteUUIDs[note.UUID] = true
	}

	// Count references FROM tags TO notes
	totalRefs := 0
	matchedRefs := 0
	refTypesSeen := make(map[string]int)
	tagsWithRefs := 0

	for _, item := range rawTags {
		tag := item.(*items.Tag)
		title := tag.Content.GetTitle()
		refs := tag.Content.References()

		noteCount := 0

		// Count Note-type references from this tag
		for _, ref := range refs {
			refTypesSeen[ref.ContentType]++

			if ref.ContentType == common.SNItemTypeNote {
				totalRefs++
				// Check if this note UUID exists in our note list
				if noteUUIDs[ref.UUID] {
					noteCount++
					matchedRefs++
				}
			}
		}

		if len(refs) > 0 {
			tagsWithRefs++
		}

		if opts.debug {
			pterm.Debug.Printf("Tag '%s': %d references, %d matched notes\n", title, len(refs), noteCount)
		}

		tagStats[tag.UUID] = &TagStats{
			Title:     title,
			UUID:      tag.UUID,
			NoteCount: noteCount,
			CreatedAt: tag.CreatedAt,
		}
	}

	// Always show summary to help diagnose
	fmt.Printf("\nDiagnostics (Tag → Note counting):\n")
	fmt.Printf("  Tags with references: %d / %d\n", tagsWithRefs, len(rawTags))
	fmt.Printf("  Reference types seen in tags: %v\n", refTypesSeen)
	fmt.Printf("  SNItemTypeNote constant = '%s'\n", common.SNItemTypeNote)
	fmt.Printf("  Total note references from tags: %d, Matched: %d, Unmatched: %d\n", totalRefs, matchedRefs, totalRefs-matchedRefs)

	if opts.debug {
		pterm.Debug.Printf("\nAdditional debug info above\n")
	}

	// Convert to slice for sorting
	var stats []*TagStats
	for _, s := range tagStats {
		stats = append(stats, s)
	}

	// Sort by note count descending
	sort.Slice(stats, func(i, j int) bool {
		return stats[i].NoteCount > stats[j].NoteCount
	})

	if len(stats) == 0 {
		pterm.Info.Println("No tags found")
		return nil
	}

	// Count tags with notes
	tagsWithNotes := 0
	for _, s := range stats {
		if s.NoteCount > 0 {
			tagsWithNotes++
		}
	}

	if opts.debug {
		pterm.Debug.Printf("Tags with notes: %d, Tags without notes: %d\n", tagsWithNotes, len(stats)-tagsWithNotes)
	}

	// Display cloud
	pterm.DefaultHeader.WithBackgroundStyle(pterm.NewStyle(pterm.BgMagenta)).
		WithMargin(10).
		Println("🏷️  Tag Cloud")
	pterm.Println()

	displayCloud(stats)

	pterm.Println()
	pterm.Info.Printf("Total: %d tag(s), %d note(s)\n", len(stats), len(rawNotes))
	if tagsWithNotes < len(stats) {
		pterm.Info.Printf("Note: %d unused tag(s) hidden from cloud view\n", len(stats)-tagsWithNotes)
	}

	return nil
}

// displayCloud renders the tag cloud
func displayCloud(stats []*TagStats) {
	if len(stats) == 0 {
		return
	}

	// Find max count for sizing
	maxCount := stats[0].NoteCount
	if maxCount == 0 {
		maxCount = 1
	}

	// Define size levels
	minSize := 1
	maxSize := 5

	// Group tags by size
	var cloudLines []string
	currentLine := ""
	lineWidth := 0
	maxLineWidth := 100

	for _, stat := range stats {
		if stat.NoteCount == 0 {
			continue
		}

		// Calculate size (1-5)
		ratio := float64(stat.NoteCount) / float64(maxCount)
		size := minSize + int(math.Round(ratio*float64(maxSize-minSize)))

		// Format tag with size and color
		tag := formatTagForCloud(stat.Title, stat.NoteCount, size)

		tagLen := len(stat.Title) + 4 // Approximate visual length

		// Check if we need a new line
		if lineWidth+tagLen > maxLineWidth && currentLine != "" {
			cloudLines = append(cloudLines, currentLine)
			currentLine = ""
			lineWidth = 0
		}

		// Add tag to current line
		if currentLine != "" {
			currentLine += "  "
			lineWidth += 2
		}
		currentLine += tag
		lineWidth += tagLen
	}

	// Add final line
	if currentLine != "" {
		cloudLines = append(cloudLines, currentLine)
	}

	// Display cloud
	for _, line := range cloudLines {
		fmt.Println(line)
	}

	// Show legend
	pterm.Println()
	pterm.DefaultSection.Println("Legend")
	pterm.Println("  Size indicates number of notes (larger = more notes)")
	pterm.Println("  Color: " + color.Red.Sprint("◼ 10+") + " " +
		color.Yellow.Sprint("◼ 5-9") + " " +
		color.Green.Sprint("◼ 3-4") + " " +
		color.Cyan.Sprint("◼ 1-2"))
}

// formatTagForCloud formats a tag for cloud display
func formatTagForCloud(title string, count int, size int) string {
	// Choose color based on count
	var colorFunc func(...interface{}) string

	switch {
	case count >= 10:
		colorFunc = color.Red.Sprint
	case count >= 5:
		colorFunc = color.Yellow.Sprint
	case count >= 3:
		colorFunc = color.Green.Sprint
	default:
		colorFunc = color.Cyan.Sprint
	}

	// Format with size
	tag := fmt.Sprintf("%s(%d)", title, count)

	// Apply size styling
	switch size {
	case 5:
		return colorFunc(strings.ToUpper(tag))
	case 4:
		return color.Bold.Sprint(colorFunc(tag))
	case 3:
		return colorFunc(tag)
	case 2:
		return colorFunc(tag)
	default:
		return color.Gray.Sprint(colorFunc(tag))
	}
}

// ShowTagStats displays detailed tag statistics
func ShowTagStats(opts configOptsOutput) error {
	// Get session
	session, _, err := cache.GetSession(common.NewHTTPClient(), opts.useSession, opts.sessKey, opts.server, opts.debug)
	if err != nil {
		return err
	}

	session.CacheDBPath, err = cache.GenCacheDBPath(session, opts.cacheDBDir, snAppName)
	if err != nil {
		return err
	}

	// Use existing sync-based approach that works for other commands
	rawTags, rawNotes, err := getItemsViaSync(&session, opts.debug)
	if err != nil {
		return err
	}

	// Build statistics - count from tags (Tag → Note references)
	tagStats := make(map[string]*TagStats)
	noteUUIDs := make(map[string]bool)

	// Build note UUID map for validation
	for _, item := range rawNotes {
		note := item.(*items.Note)
		noteUUIDs[note.UUID] = true
	}

	// Count references FROM tags TO notes
	for _, item := range rawTags {
		tag := item.(*items.Tag)
		refs := tag.Content.References()

		noteCount := 0

		// Count Note-type references from this tag
		for _, ref := range refs {
			if ref.ContentType == common.SNItemTypeNote {
				// Check if this note UUID exists in our note list
				if noteUUIDs[ref.UUID] {
					noteCount++
				}
			}
		}

		tagStats[tag.UUID] = &TagStats{
			Title:     tag.Content.GetTitle(),
			UUID:      tag.UUID,
			NoteCount: noteCount,
			CreatedAt: tag.CreatedAt,
		}
	}

	// Convert to slice and sort
	var stats []*TagStats
	for _, s := range tagStats {
		stats = append(stats, s)
	}

	sort.Slice(stats, func(i, j int) bool {
		return stats[i].NoteCount > stats[j].NoteCount
	})

	// Display table
	pterm.DefaultSection.Println("Tag Statistics")

	tableData := [][]string{
		{color.Cyan.Sprint("#"), color.Cyan.Sprint("Tag"), color.Cyan.Sprint("Notes"), color.Cyan.Sprint("Created")},
	}

	for i, stat := range stats {
		noteCount := fmt.Sprintf("%d", stat.NoteCount)
		if stat.NoteCount == 0 {
			noteCount = color.Gray.Sprint("0")
		} else if stat.NoteCount >= 10 {
			noteCount = color.Red.Sprint(noteCount)
		} else if stat.NoteCount >= 5 {
			noteCount = color.Yellow.Sprint(noteCount)
		} else {
			noteCount = color.Green.Sprint(noteCount)
		}

		created := ""
		if len(stat.CreatedAt) >= 10 {
			created = stat.CreatedAt[:10]
		}

		tableData = append(tableData, []string{
			color.Gray.Sprint(fmt.Sprintf("%d", i+1)),
			stat.Title,
			noteCount,
			color.Gray.Sprint(created),
		})
	}

	pterm.DefaultTable.WithHasHeader(true).
		WithHeaderStyle(pterm.NewStyle(pterm.FgLightCyan, pterm.Bold)).
		WithData(tableData).
		WithBoxed(true).
		Render()

	// Show summary
	pterm.Println()
	totalNotes := len(rawNotes)
	avgNotesPerTag := 0.0
	if len(stats) > 0 {
		totalTaggedNotes := 0
		for _, s := range stats {
			totalTaggedNotes += s.NoteCount
		}
		avgNotesPerTag = float64(totalTaggedNotes) / float64(len(stats))
	}

	pterm.Info.Printf("Total: %d tag(s), %d note(s)\n", len(stats), totalNotes)
	pterm.Info.Printf("Average: %.1f notes per tag\n", avgNotesPerTag)

	// Find unused tags
	unusedCount := 0
	for _, s := range stats {
		if s.NoteCount == 0 {
			unusedCount++
		}
	}
	if unusedCount > 0 {
		pterm.Warning.Printf("%d unused tag(s)\n", unusedCount)
	}

	return nil
}
