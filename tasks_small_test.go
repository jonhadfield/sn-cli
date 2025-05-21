package sncli

import (
	"testing"

	"github.com/jonhadfield/gosn-v2/items"
	"github.com/stretchr/testify/require"
)

func TestBoolToText(t *testing.T) {
	require.Equal(t, "yes", boolToText(true, "yes", "no"))
	require.Equal(t, "no", boolToText(false, "yes", "no"))
}

func TestOutputChars(t *testing.T) {
	require.Equal(t, "hello", outputChars("hello", 10))
	long := "abcdefghijklmnop"
	require.Equal(t, "abcdefghij...", outputChars(long, 10))
}

func TestTaskListsConflictedWarning(t *testing.T) {
	tasks := []items.Tasklist{{}, {}}
	require.Contains(t, taskListsConflictedWarning(tasks), "2 conflicted versions")
	require.Equal(t, "-", taskListsConflictedWarning(nil))
}

func TestFilterTasks(t *testing.T) {
	tasks := items.Tasks{
		{Title: "done", Completed: true},
		{Title: "todo", Completed: false},
	}
	completed := filterTasks(tasks, true)
	require.Len(t, completed, 1)
	require.Equal(t, "done", completed[0].Title)
	incomplete := filterTasks(tasks, false)
	require.Len(t, incomplete, 1)
	require.Equal(t, "todo", incomplete[0].Title)
}

func TestFilterAdvancedChecklistTasks(t *testing.T) {
	tasks := items.AdvancedChecklistTasks{
		{Description: "d1", Completed: true},
		{Description: "d2", Completed: false},
	}
	completed := filterAdvancedChecklistTasks(tasks, true)
	require.Len(t, completed, 1)
	require.Equal(t, "d1", completed[0].Description)
	incomplete := filterAdvancedChecklistTasks(tasks, false)
	require.Len(t, incomplete, 1)
	require.Equal(t, "d2", incomplete[0].Description)
}
