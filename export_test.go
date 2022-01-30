package sncli

import (
	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecryptWithNewItemsKey(t *testing.T) {
	ik, err := testSession.CreateItemsKey()
	require.NoError(t, err)
	require.Equal(t, "SN|ItemsKey", ik.ContentType)
	require.False(t, ik.Deleted)
	require.NotEmpty(t, ik.UUID)
	require.NotEmpty(t, ik.Content)
	require.NotEmpty(t, ik.CreatedAt)
	require.NotEmpty(t, ik.CreatedAtTimestamp)
	require.Empty(t, ik.UpdatedAtTimestamp)
	require.Empty(t, ik.UpdatedAt)

	n := gosn.NewNote()
	nc := gosn.NewNoteContent()
	nc.Title = "test title"
	nc.Text = "test content"
	n.Content = *nc

	eis := gosn.Items{&n}
	encItems, err := eis.Encrypt(ik, testSession.MasterKey, testSession.Debug)
	require.NoError(t, err)
	require.NotEmpty(t, encItems[0].UUID)
	require.NotEmpty(t, encItems[0].CreatedAtTimestamp)
	require.NotEmpty(t, encItems[0].CreatedAt)
	require.False(t, encItems[0].Deleted)
	require.Equal(t, "Note", encItems[0].ContentType)
	require.NotEmpty(t, encItems[0].Content)
	require.Empty(t, encItems[0].DuplicateOf)

	testSession.ItemsKeys = append(testSession.Session.ItemsKeys, ik)
	di, err := encItems.Decrypt(testSession.Session, gosn.ItemsKey{})

	require.NoError(t, err)
	require.NotEmpty(t, di)
	require.Greater(t, len(di), 0)

	pi, err := encItems.DecryptAndParse(testSession.Session)
	require.NoError(t, err)
	require.Greater(t, len(pi), 0)

	var dn gosn.Note

	for x := range pi {
		if pi[x].GetContentType() == "Note" {
			dn = *pi[x].(*gosn.Note)

		}
	}

	require.Equal(t, "Note", dn.ContentType)
	require.Equal(t, "test title", dn.Content.GetTitle())
	require.Equal(t, "test content", dn.Content.GetText())
	require.NotEmpty(t, dn.UUID)
	require.NotEmpty(t, dn.CreatedAtTimestamp)
	require.NotEmpty(t, dn.CreatedAt)
	require.False(t, dn.Deleted)

}

func TestEncryptionDecryptionOfItemsKey(t *testing.T) {
	ik, err := testSession.CreateItemsKey()
	require.NoError(t, err)
	require.Equal(t, "SN|ItemsKey", ik.ContentType)
	require.False(t, ik.Deleted)
	require.NotEmpty(t, ik.UUID)
	require.NotEmpty(t, ik.Content)
	require.NotEmpty(t, ik.CreatedAt)
	require.NotEmpty(t, ik.CreatedAtTimestamp)
	require.Empty(t, ik.UpdatedAtTimestamp)
	require.Empty(t, ik.UpdatedAt)

	newItemsKeyUUID := ik.UUID
	newItemsKeyItemsKey := ik.ItemsKey

	eik, err := ik.Encrypt(testSession.Session, true)
	require.NoError(t, err)
	dik, err := gosn.DecryptAndParseItemKeys(testSession.Session.MasterKey, []gosn.EncryptedItem{eik})
	require.NoError(t, err)
	require.Equal(t, newItemsKeyUUID, dik[0].UUID)
	require.Equal(t, newItemsKeyItemsKey, dik[0].ItemsKey)
}

// export and import of a note to check encryption and cache propagation
func TestJSONExportImport(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// populate DB
	gii := cache.SyncInput{
		Session: testSession,
	}

	gio, err := Sync(gii, false)
	require.NoError(t, err)
	// DB now populated and open with pointer in session
	var existingItems []cache.Item
	err = gio.DB.All(&existingItems)
	require.NoError(t, err)
	// record number of items before adding the note for later comparison
	allCacheItemsPreExport := len(existingItems)
	allItemsKeysPreExport := len(testSession.ItemsKeys)

	note := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	note.Content = *noteContent
	note.Content.SetTitle("Example Title")
	note.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&note,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	require.NoError(t, err)
	require.Len(t, encItemsToPut, 1)

	cItems := cache.ToCacheItems(encItemsToPut, false)
	for _, cItem := range cItems {
		require.NoError(t, gio.DB.Save(&cItem))
	}
	require.Len(t, cItems, 1)

	require.NoError(t, gio.DB.Close())

	pii := cache.SyncInput{
		Session: testSession,
	}

	// sync note (in cache) to SN
	var so cache.SyncOutput
	so, err = Sync(pii, false)
	require.NoError(t, err)

	require.NoError(t, so.DB.Close())

	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)

	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			panic("failed to remove temp dir")
		}
	}() // clean up

	tmpfn := filepath.Join(dir, "tmpfile")

	// copy default items key before export for later comparison
	defaultItemsKeyPreExport := testSession.DefaultItemsKey.ItemsKey
	err = testSession.Export(tmpfn)
	require.NoError(t, err)

	// ensure default items key has not changed since export (export will contain new items key)
	require.Equal(t, defaultItemsKeyPreExport, testSession.DefaultItemsKey.ItemsKey)

	err = testSession.Import(tmpfn, true)
	require.NoError(t, err)
	// ensure default items key has changed after import, as all items should now be re-encrypted with new key
	require.NotEqual(t, defaultItemsKeyPreExport, testSession.DefaultItemsKey.ItemsKey)

	// get a new database and populate with the new item
	gii = cache.SyncInput{
		Session: testSession,
	}

	require.NoError(t, gio.DB.Close())
	require.NoError(t, gii.Session.CacheDB.Close())
	gio, err = Sync(gii, false)
	require.NoError(t, err)

	require.NotNil(t, gio)
	require.NotEmpty(t, gio.DB)

	var aa []cache.Item

	require.NoError(t, gio.DB.All(&aa))
	require.Equal(t, allCacheItemsPreExport+1, len(aa))
	var importedNote cache.Item

	for _, i := range aa {
		if i.ContentType == "Note" {
			if i.UUID == note.UUID {
				importedNote = i
			}
		}
	}

	require.NotNil(t, importedNote.ItemsKeyID)
	require.Equal(t, allItemsKeysPreExport, len(testSession.ItemsKeys))

	require.NoError(t, gio.DB.Close())
}

// export one note, delete that note, import the backup and check note has returned.
func TestJSONExportWipeImportOneNote(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// populate DB
	gii := cache.SyncInput{
		Session: testSession,
	}

	gio, err := Sync(gii, false)
	require.NoError(t, err)
	// DB now populated and open with pointer in session
	var existingItems []cache.Item
	err = gio.DB.All(&existingItems)
	require.NoError(t, err)

	note := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	note.Content = *noteContent
	note.Content.SetTitle("Example Title")
	note.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&note,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	require.NoError(t, err)

	cItems := cache.ToCacheItems(encItemsToPut, false)
	for _, cItem := range cItems {
		require.NoError(t, gio.DB.Save(&cItem))
	}

	require.NoError(t, gio.DB.Close())

	pii := cache.SyncInput{
		Session: testSession,
	}

	var so cache.SyncOutput
	so, err = Sync(pii, false)
	require.NoError(t, err)
	require.NoError(t, so.DB.Close())

	dir, err := ioutil.TempDir("", "test")
	require.NoError(t, err)

	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			panic("failed to remove temp dir")
		}
	}() // clean up

	tmpfn := filepath.Join(dir, "tmpfile")

	err = testSession.Export(tmpfn)
	require.NoError(t, err)

	// delete note
	so, err = Sync(cache.SyncInput{
		Session: testSession,
		Close:   false,
	}, true)

	var preImportItems []cache.Item
	require.NoError(t, so.DB.All(&preImportItems))
	var piNote cache.Item
	for x := range preImportItems {
		if preImportItems[x].UUID == note.UUID {
			piNote = preImportItems[x]
			piNote.Deleted = true
			require.NoError(t, so.DB.Save(&piNote))
		}
	}
	// resync to remove note
	require.NotNil(t, piNote.ItemsKeyID)
	require.NoError(t, so.DB.Close())
	so, err = Sync(cache.SyncInput{
		Session: testSession,
		Close:   true,
	}, true)

	err = testSession.Import(tmpfn, true)
	require.NoError(t, err)

	// get a new database and populate with the new item
	gii = cache.SyncInput{
		Session: testSession,
		Close:   false,
	}

	gio, err = Sync(gii, false)
	require.NoError(t, err)

	var aa []cache.Item

	require.NoError(t, gio.DB.All(&aa))

	var foundNote cache.Item

	for _, i := range aa {
		if i.ContentType == "Note" {
			if i.UUID == note.UUID {
				require.False(t, i.Deleted)
				foundNote = i
			}
		}
	}

	require.NotEmpty(t, foundNote.ItemsKeyID)
	require.NoError(t, gio.DB.Close())

	// decrypt note
	citd := cache.Items{foundNote}

	itd, err := citd.ToItems(testSession)
	require.NoError(t, err)
	require.Len(t, itd, 1)
	dn := itd[0].(*gosn.Note)
	require.Equal(t, dn.Content.Title, note.Content.Title)
	require.Equal(t, dn.Content.Text, note.Content.Text)
}

// Create a note, export it, change original, import and check a duplicate has been created.
func TestConflictResolution(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// create and put initial originalNote
	originalNote := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	originalNote.Content = *noteContent
	originalNoteTitle := "Example Title"
	originalNoteText := "Some example text"

	originalNote.Content.SetTitle(originalNoteTitle)
	originalNote.Content.SetText(originalNoteText)

	itemsToPut := gosn.Items{
		&originalNote,
	}

	encItemsToPut, err := itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	assert.NoError(t, err)

	// perform initial sync to load keys into session
	pii := cache.SyncInput{
		Session: testSession,
	}

	var so cache.SyncOutput
	so, err = Sync(pii, false)

	require.NoError(t, err)

	pi := cache.ToCacheItems(encItemsToPut, false)

	require.Len(t, itemsToPut, 1)

	for _, p := range pi {
		assert.NoError(t, so.DB.Save(&p))
	}

	assert.NoError(t, so.DB.Close())

	// sync saved item in db to SN
	so, err = Sync(pii, false)
	assert.NoError(t, err)
	assert.NoError(t, so.DB.Close())

	// get db
	pii.Close = false
	so, err = Sync(pii, false)
	assert.NoError(t, err)
	// Get items in DB to see what's in there
	var encItems cache.Items
	err = so.DB.All(&encItems)
	assert.NoError(t, err)

	// change initial originalNote and re-put
	updatedNote := originalNote.Copy()
	updatedNote.Content.SetTitle("Example Title UPDATED")
	updatedNote.Content.SetText("Some example text UPDATED")

	itemsToPut = gosn.Items{
		&updatedNote,
	}

	encItemsToPut, err = itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	assert.NoError(t, err)
	pi = cache.ToCacheItems(encItemsToPut, false)
	for _, i := range pi {
		assert.NoError(t, so.DB.Save(&i))
	}

	assert.NoError(t, so.DB.Close())
	so, err = Sync(pii, false)
	require.NoError(t, err)

	var final cache.Items
	err = so.DB.All(&final)

	var origFound bool

	var newItemWithDupeIDBeingOrig bool

	for _, x := range final {
		if x.UUID == originalNote.GetUUID() {
			origFound = true
		}

		if x.ContentType == "Note" {
			if x.UUID != originalNote.UUID {
				newItemWithDupeIDBeingOrig = true
			}
		}
	}

	require.True(t, origFound)
	require.True(t, newItemWithDupeIDBeingOrig)
	require.NoError(t, so.DB.Close())
}

func TestExportChangeImportOneTag(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// create and put initial originalTag
	originalTag := gosn.NewTag()
	tagContent := gosn.NewTagContent()
	originalTag.Content = *tagContent
	originalTag.Content.SetTitle("Example Title")
	itemsToPut := gosn.Items{
		&originalTag,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	assert.NoError(t, err)

	// get db
	pii := cache.SyncInput{
		Session: testSession,
	}

	var so cache.SyncOutput
	so, err = Sync(pii, false)

	assert.NoError(t, err)

	// add item to db
	ci := cache.ToCacheItems(encItemsToPut, false)
	for _, i := range ci {
		assert.NoError(t, so.DB.Save(&i))
	}

	assert.NoError(t, so.DB.Close())

	// sync db with SN
	so, err = Sync(pii, false)
	assert.NoError(t, err)
	// close db
	assert.NoError(t, so.DB.Close())

	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)

	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			panic("failed to remove temp dir")
		}
	}() // clean up

	// export initial originalTag
	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if ecErr := ec.Run(); ecErr != nil {
		panic(ecErr)
	}

	// change initial originalTag and re-put
	updatedTag := originalTag.Copy()
	updatedTag.Content.SetTitle("Example Title UPDATED")
	itemsToPut = gosn.Items{
		&updatedTag,
	}
	encItemsToPut, err = itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)

	assert.NoError(t, err)

	// get db
	so, err = Sync(pii, false)
	assert.NoError(t, err)
	// add items to db
	ci = cache.ToCacheItems(encItemsToPut, false)
	for _, i := range ci {
		assert.NoError(t, so.DB.Save(&i))
	}

	assert.NoError(t, so.DB.Close())

	pii = cache.SyncInput{
		Session: testSession,
		Close:   true,
	}
	_, err = Sync(pii, false)
	assert.NoError(t, err)

	// import original export
	ic := ImportConfig{
		Session: testSession,
		Format:  "json",
		File:    tmpfn,
	}

	_, err = ic.Run()
	assert.NoError(t, err)

	// get items again
	gii := cache.SyncInput{
		Session: testSession,
	}

	var gio cache.SyncOutput
	gio, err = Sync(gii, false)
	assert.NoError(t, err)

	var cItems cache.Items

	assert.NoError(t, gio.DB.All(&cItems))
	assert.NoError(t, gio.DB.Close())

	var gItems gosn.Items
	gItems, err = cItems.ToItems(testSession)

	assert.NoError(t, err)

	var found bool

	for _, i := range gItems {
		if i.GetContentType() == "Tag" {
			if i.(*gosn.Tag).Equals(originalTag) {
				found = true
			}
		}
	}

	assert.True(t, found)
}

func TestExportDeleteImportOneTag(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	pii := cache.SyncInput{
		Session: testSession,
	}

	// Get DB
	so, err := Sync(pii, false)
	assert.NoError(t, err)

	// create and put originalTag
	originalTag := gosn.NewTag()
	tagContent := gosn.NewTagContent()
	originalTag.Content = *tagContent
	originalTag.Content.SetTitle("Example Title")
	itemsToPut := gosn.Items{
		&originalTag,
	}

	encItemsToPut, err := itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	assert.NoError(t, err)

	if err = cache.SaveEncryptedItems(so.DB, encItemsToPut, true); err != nil {
		return
	}

	var cItems cache.Items

	so, err = Sync(pii, false)
	assert.NoError(t, err)
	assert.NoError(t, so.DB.Close())

	// Export existing content to a temporary directory
	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)

	defer func() {
		if err = os.RemoveAll(dir); err != nil {
			panic("failed to remove temp dir")
		}
	}() // clean up

	// export initial originalTag
	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if ecErr := ec.Run(); ecErr != nil {
		panic(ecErr)
	}

	so, err = Sync(cache.SyncInput{
		Session: testSession,
		Close:   false,
	}, true)

	err = so.DB.All(&cItems)
	require.NoError(t, err)
	require.NoError(t, so.DB.Close())
	var cTagToDel cache.Item
	for _, cItem := range cItems {
		if originalTag.UUID == cItem.UUID {
			cTagToDel = cItem
		}
	}
	require.NotEmpty(t, cTagToDel.UUID)

	// get original tag from db so we can delete
	// delete originalTag
	eItems := cache.Items{cTagToDel}
	dItems, err := eItems.ToItems(testSession)
	require.NoError(t, err)
	require.Len(t, dItems, 1)
	itd := dItems[0]
	itd.SetDeleted(true)

	itemsToPut = gosn.Items{
		itd,
	}

	encItemsToPut, err = itemsToPut.Encrypt(testSession.Session.DefaultItemsKey, testSession.Session.MasterKey, testSession.Session.Debug)
	assert.NoError(t, err)

	so, err = Sync(pii, false)
	assert.NoError(t, err)

	if err = cache.SaveEncryptedItems(so.DB, encItemsToPut, true); err != nil {
		return
	}

	pii = cache.SyncInput{
		Session: testSession,
	}
	so, err = Sync(pii, false)
	assert.NoError(t, err)
	assert.NoError(t, so.DB.Close())

	// import original export
	ic := ImportConfig{
		Session: testSession,
		Format:  "json",
		File:    tmpfn,
	}

	_, err = ic.Run()
	assert.NoError(t, err)

	// get items again
	gii := cache.SyncInput{
		Session: testSession,
	}

	var gio cache.SyncOutput
	gio, err = Sync(gii, false)
	assert.NoError(t, err)

	assert.NoError(t, gio.DB.All(&cItems))

	assert.NoError(t, gio.DB.Close())

	var gItems gosn.Items
	gItems, err = cItems.ToItems(testSession)

	assert.NoError(t, err)

	var found bool

	for _, i := range gItems {
		if i.GetContentType() == "Tag" {
			if i.GetUUID() == originalTag.UUID {
				found = true
			}
		}
	}

	assert.True(t, found)
}
