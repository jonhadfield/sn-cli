package sncli

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/jonhadfield/gosn-v2"
	"github.com/stretchr/testify/assert"
)

func TestExportOneNote(t *testing.T) {
	cleanUp(&testSession)
	defer cleanUp(&testSession)

	note := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	note.Content = *noteContent
	note.Content.SetTitle("Example Title")
	note.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&note,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii := gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)
	}

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
	writtenItems, err = writtenEncryptedItems.DecryptAndParse(testSession.Mk, testSession.Ak, true)
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

func TestExportWipeImportOneNote(t *testing.T) {
	defer cleanUp(&testSession)

	note := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	note.Content = *noteContent
	note.Content.SetTitle("Example Title")
	note.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&note,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii := gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(dir) // clean up

	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if ecErr := ec.Run(); ecErr != nil {
		panic(ecErr)
	}

	cleanUp(&testSession)

	ic := ImportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if icErr := ic.Run(); icErr != nil {
		panic(icErr)
	}

	gii := gosn.SyncInput{
		Session: testSession,
	}

	var gio gosn.SyncOutput
	gio, err = gosn.Sync(gii)
	assert.NoError(t, err)

	gio.Items = filterByTypes(gio.Items, supportedContentTypes)


	var items gosn.Items
	items, err = gio.Items.DecryptAndParse(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	var found bool

	for _, i := range items {
		switch i.(type) {
		case *gosn.Note:
			if i.(*gosn.Note).Equals(note) {
				found = true
			}
		}
	}

	assert.True(t, found)
}

func TestExportChangeImportOneNote(t *testing.T) {
	defer cleanUp(&testSession)

	// create and put initial originalNote
	originalNote := gosn.NewNote()
	noteContent := gosn.NewNoteContent()
	originalNote.Content = *noteContent
	originalNote.Content.SetTitle("Example Title")
	originalNote.Content.SetText("Some example text")
	itemsToPut := gosn.Items{
		&originalNote,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii := gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir("", "test")
	assert.NoError(t, err)

	defer os.RemoveAll(dir) // clean up

	// export initial originalNote
	tmpfn := filepath.Join(dir, "tmpfile")
	ec := ExportConfig{
		Session: testSession,
		File:    tmpfn,
	}
	err = ec.Run()
	assert.NoError(t, err)

	// change initial originalNote and re-put
	updatedNote := originalNote.Copy()
	updatedNote.Content.SetTitle("Example Title UPDATED")
	updatedNote.Content.SetText("Some example text UPDATED")
	itemsToPut = gosn.Items{
		&updatedNote,
	}
	encItemsToPut, err = itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii = gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	// import original export
	ic := ImportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	if icErr := ic.Run(); icErr != nil {
		panic(icErr)
	}

	// get items again
	gii := gosn.SyncInput{
		Session: testSession,
	}

	var gio gosn.SyncOutput
	gio, err = gosn.Sync(gii)
	assert.NoError(t, err)

	var items gosn.Items
	items, err = gio.Items.DecryptAndParse(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	var found bool

	for _, i := range items {
		switch i.(type) {
		case *gosn.Note:
			if i.(*gosn.Note).Equals(originalNote) {
				found = true
			}
		}
	}

	assert.True(t, found)
}

func TestExportChangeImportOneTag(t *testing.T) {
	defer cleanUp(&testSession)

	// create and put initial originalTag
	originalTag := gosn.NewTag()
	tagContent := gosn.NewTagContent()
	originalTag.Content = *tagContent
	originalTag.Content.SetTitle("Example Title")
	itemsToPut := gosn.Items{
		&originalTag,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii := gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)
	}

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
	encItemsToPut, err = itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii = gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	// import original export
	ic := ImportConfig{
		Session: testSession,
		File:    tmpfn,
	}

	err = ic.Run()
	assert.NoError(t, err)

	// get items again
	gii := gosn.SyncInput{
		Session: testSession,
	}

	var gio gosn.SyncOutput
	gio, err = gosn.Sync(gii)
	assert.NoError(t, err)

	var items gosn.Items
	items, err = gio.Items.DecryptAndParse(testSession.Mk, testSession.Ak, true)

	assert.NoError(t, err)

	var found bool

	for _, i := range items {
		switch i.(type) {
		case *gosn.Tag:
			if i.(*gosn.Tag).Equals(originalTag) {
				found = true
			}
		}
	}

	assert.True(t, found)
}

func TestExportDeleteImportOneTag(t *testing.T) {
	defer cleanUp(&testSession)
	// create and put originalTag
	originalTag := gosn.NewTag()
	tagContent := gosn.NewTagContent()
	originalTag.Content = *tagContent
	originalTag.Content.SetTitle("Example Title")
	itemsToPut := gosn.Items{
		&originalTag,
	}
	encItemsToPut, err := itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii := gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)

	dir, err := ioutil.TempDir("", "test")
	if err != nil {
		log.Fatal(err)
	}

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

	// create copy of original tag
	copyOfOriginalTag := originalTag.Copy()
	// delete originalTag
	originalTag.Deleted = true
	itemsToPut = gosn.Items{
		&originalTag,
	}
	encItemsToPut, err = itemsToPut.Encrypt(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	pii = gosn.SyncInput{
		Session: testSession,
		Items:   encItemsToPut,
	}
	_, err = gosn.Sync(pii)
	assert.NoError(t, err)
	// import original export
	ic := ImportConfig{
		Session: testSession,
		File:    tmpfn,
	}
	err = ic.Run()
	assert.NoError(t, err)

	// get items again
	gii := gosn.SyncInput{
		Session: testSession,
	}

	var gio gosn.SyncOutput
	gio, err = gosn.Sync(gii)
	assert.NoError(t, err)

	var items gosn.Items
	items, err = gio.Items.DecryptAndParse(testSession.Mk, testSession.Ak, true)
	assert.NoError(t, err)

	var found bool

	for _, i := range items {
		switch i.(type) {
		case *gosn.Tag:
			if i.(*gosn.Tag).Equals(copyOfOriginalTag) {
				found = true
			}
		}
	}

	assert.True(t, found)
}
