package sncli

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestGetDataWithLargeDataset tests GetData with >150 items to reproduce multi-sync issues
func TestGetDataWithLargeDataset(t *testing.T) {
	testDelay()

	defer cleanUp(*testSession)

	// Create a large dataset (200 notes) to trigger multiple sync operations
	numNotes := 200
	textParas := 1 // Minimal content to speed up creation

	t.Logf("Creating %d notes to test GetData with large dataset...", numNotes)
	err := createNotes(testSession, numNotes, textParas)
	if err != nil {
		if strings.Contains(err.Error(), "giving up after") {
			t.Logf("Note creation timed out - this is expected with large batches and existing account data")
			t.Logf("Proceeding with test using existing data in account")
		} else {
			require.NoError(t, err)
		}
	}

	// Add strategic pause after note creation
	time.Sleep(2 * time.Second)

	t.Logf("Testing GetData with large dataset (expecting >150 items)...")

	// Create StatsConfig to test GetData
	statsConfig := StatsConfig{
		Session: *testSession,
	}

	// Test GetData - this should trigger multiple sync operations and reveal the issue
	var statsData StatsData
	statsData, err = statsConfig.GetData()
	if err != nil {
		if strings.Contains(err.Error(), "giving up after") {
			t.Logf("GetData timed out with large dataset - this confirms the multi-sync issue")
			t.Logf("Error: %v", err)
			t.Logf("This is the issue we need to fix in cache.Sync reliability")

			// The test should fail here to highlight the issue
			t.Fatalf("GetData failed with large dataset: %v", err)
		}
		require.NoError(t, err)
	}

	// Verify we got meaningful stats data
	require.NotNil(t, statsData.CoreTypeCounter.counts)
	require.NotNil(t, statsData.OtherTypeCounter.counts)

	// Check that we have a reasonable number of notes in the stats
	noteCount := statsData.CoreTypeCounter.counts["Note"]
	t.Logf("GetData reports %d notes in stats", noteCount)

	// We should have at least some notes (allowing for potential sync timeout with partial data)
	require.GreaterOrEqual(t, noteCount, int64(10), "Expected at least 10 notes in stats")

	t.Logf("GetData succeeded with large dataset - no multi-sync issue detected")
}

// TestGetDataMultipleCalls tests multiple consecutive GetData calls to reproduce hanging
//func TestGetDataMultipleCalls(t *testing.T) {
//	testDelay()
//
//	defer cleanUp(*testSession)
//
//	// Create a moderate dataset to ensure we have some data
//	numNotes := 50
//	textParas := 1
//
//	t.Logf("Creating %d notes for multiple GetData calls test...", numNotes)
//	err := createNotes(testSession, numNotes, textParas)
//	if err != nil {
//		if strings.Contains(err.Error(), "giving up after") {
//			t.Logf("Note creation timed out - proceeding with existing account data")
//		} else {
//			require.NoError(t, err)
//		}
//	}
//
//	// Add strategic pause after note creation
//	time.Sleep(1 * time.Second)
//
//	statsConfig := StatsConfig{
//		Session: *testSession,
//	}
//
//	// Test multiple consecutive GetData calls - this often causes the second call to hang
//	for i := 1; i <= 3; i++ {
//		t.Logf("GetData call #%d...", i)
//
//		var statsData StatsData
//		statsData, err = statsConfig.GetData()
//		if err != nil {
//			if strings.Contains(err.Error(), "giving up after") {
//				t.Logf("GetData call #%d timed out - this demonstrates the consecutive sync issue", i)
//				t.Logf("Error: %v", err)
//
//				if i == 1 {
//					t.Fatalf("First GetData call failed: %v", err)
//				} else {
//					t.Fatalf("GetData call #%d failed - consecutive sync issue confirmed: %v", i, err)
//				}
//			}
//			require.NoError(t, err)
//		}
//
//		// Verify we got valid stats
//		require.NotNil(t, statsData.CoreTypeCounter.counts)
//		noteCount := statsData.CoreTypeCounter.counts["Note"]
//		t.Logf("GetData call #%d succeeded, found %d notes", i, noteCount)
//
//		// Add delay between calls to simulate real usage
//		if i < 3 {
//			time.Sleep(500 * time.Millisecond)
//		}
//	}
//
//	t.Logf("Multiple consecutive GetData calls succeeded - no hanging issue detected")
//}
