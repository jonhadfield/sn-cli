package sncli

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/jonhadfield/gosn-v2"
	"github.com/jonhadfield/gosn-v2/cache"
	"github.com/stretchr/testify/assert"
)

func TestExportOneNote(t *testing.T) {
	testDelay()

	cleanUp(*testSession)
	defer cleanUp(*testSession)

	// populate DB
	si := cache.SyncInput{
		Session: testSession,
	}

	so, err := Sync(si, false)
	assert.NoError(t, err)

	// create a note
	note := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	note.Content = *noteContent
	note.Content.SetTitle("Example Title")
	note.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&note,
	}

	encItemsToPut, err := itemsToPut.Encrypt(*testSession.Session)
	assert.NoError(t, err)

	cItems := cache.ToCacheItems(encItemsToPut, false)
	for _, ci := range cItems {
		assert.NoError(t, so.DB.Save(&ci))
	}

	assert.NoError(t, so.DB.Close())

	so, err = Sync(si, false)
	assert.NoError(t, err)
	assert.NoError(t, so.DB.Close())

	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)

	defer os.RemoveAll(dir) // clean up

	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if runErr := ec.Run(); runErr != nil {
		panic(runErr)
	}

	var writtenEncryptedItems gosn.EncryptedItems
	if expErr := readGob(tmpfn, &writtenEncryptedItems); expErr != nil {
		panic(expErr)
	}

	var writtenItems gosn.Items
	writtenItems, err = writtenEncryptedItems.DecryptAndParse(testSession.Session)
	assert.NoError(t, err)

	var found bool

	for _, item := range writtenItems {
		if item != nil && item.GetUUID() == note.UUID {
			found = true
			break
		}
	}

	assert.True(t, found)
}

// export one note, delete that note, import the backup and check note has returned.
func TestExportWipeImportOneNote(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// populate DB
	gii := cache.SyncInput{
		Session: testSession,
	}

	gio, err := Sync(gii, false)
	assert.NoError(t, err)
	// DB now populated and open with pointer in session
	var existingItems []cache.Item
	err = gio.DB.All(&existingItems)
	assert.NoError(t, err)

	note := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	note.Content = *noteContent
	note.Content.SetTitle("Example Title")
	note.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&note,
	}

	encItemsToPut, err := itemsToPut.Encrypt(*testSession.Session)
	assert.NoError(t, err)

	cItems := cache.ToCacheItems(encItemsToPut, false)
	for _, cItem := range cItems {
		assert.NoError(t, gio.DB.Save(&cItem))
	}

	assert.NoError(t, gio.DB.Close())

	pii := cache.SyncInput{
		Session: testSession,
	}

	var so cache.SyncOutput
	so, err = Sync(pii, false)

	assert.NoError(t, err)
	assert.NoError(t, so.DB.Close())

	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)

	defer os.RemoveAll(dir) // clean up

	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	assert.NoError(t, ec.Run())
	// delete the db and wipe SN
	cleanUp(*testSession)

	// import the export made above so that SN is now populated
	ic := ImportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	assert.NoError(t, ic.Run())

	// get a new database and populate with the new item
	gii = cache.SyncInput{
		Session: testSession,
	}

	assert.NoError(t, gio.DB.Close())
	gio, err = Sync(gii, false)
	assert.NoError(t, err)

	assert.NotNil(t, gio)
	assert.NotEmpty(t, gio.DB)

	var aa []cache.Item

	assert.NoError(t, gio.DB.All(&aa))

	var found bool

	for _, i := range aa {
		if i.ContentType == "Note" {
			if i.UUID == note.UUID {
				found = true
			}
		}
	}

	assert.True(t, found)

	assert.NoError(t, gio.DB.Close())
}

// Create a note, export it, change original, import and check exported items replace modified.
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

	encItemsToPut, err := itemsToPut.Encrypt(*testSession.Session)
	assert.NoError(t, err)

	// ### sync db
	pii := cache.SyncInput{
		Session: testSession,
	}

	var so cache.SyncOutput
	so, err = Sync(pii, false)

	assert.NoError(t, err)

	pi := cache.ToCacheItems(encItemsToPut, false)

	assert.NotEqual(t, 1, itemsToPut)

	for _, p := range pi {
		assert.NoError(t, so.DB.Save(&p))
	}

	assert.NoError(t, so.DB.Close())

	so, err = Sync(pii, false)
	assert.NoError(t, err)

	var encItems1 cache.Items

	err = so.DB.All(&encItems1)
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

	encItemsToPut, err = itemsToPut.Encrypt(*testSession.Session)
	assert.NoError(t, err)

	pi = cache.ToCacheItems(encItemsToPut, false)
	for _, i := range pi {
		assert.NoError(t, so.DB.Save(&i))
	}

	err = so.DB.All(&encItems)
	assert.NoError(t, err)
	assert.NoError(t, so.DB.Close())

	pii.CacheDB = nil

	assert.NoError(t, so.DB.Close())

	so, err = Sync(pii, false)
	assert.NoError(t, err)

	var final cache.Items
	err = so.DB.All(&final)

	var origFound bool

	var newItemWithDupeIDBeingOrig bool

	for _, x := range final {
		if x.UUID == originalNote.GetUUID() {
			origFound = true
		}

		if x.DuplicateOf == originalNote.GetUUID() && x.UUID != originalNote.GetUUID() {
			newItemWithDupeIDBeingOrig = true
		}
	}

	assert.True(t, origFound)
	assert.True(t, newItemWithDupeIDBeingOrig)
	assert.NoError(t, so.DB.Close())
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
	encItemsToPut, err := itemsToPut.Encrypt(*testSession.Session)
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

	defer os.RemoveAll(dir) // clean up

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
	encItemsToPut, err = itemsToPut.Encrypt(*testSession.Session)
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
		File:    tmpfn,
	}

	err = ic.Run()
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
	encItemsToPut, err := itemsToPut.Encrypt(*testSession.Session)
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

	defer os.RemoveAll(dir) // clean up

	// export initial originalTag
	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if ecErr := ec.Run(); ecErr != nil {
		panic(ecErr)
	}

	// delete originalTag
	originalTag.Deleted = true
	itemsToPut = gosn.Items{
		&originalTag,
	}
	encItemsToPut, err = itemsToPut.Encrypt(*testSession.Session)
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
		File:    tmpfn,
	}

	err = ic.Run()
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
