package client

import (
	"testing"
)

func TestClientHistoryExtended(t *testing.T) {
	c, err := FullClientTestEnvironment(t.Name())
	if err != nil {
		t.Fatal(err)
	}

	// Initial check: All entries should be empty.
	for i := uint32(0); i < 100; i++ {
		amt, err := c.loadReading(i)
		if err != nil {
			t.Fatal(err)
		}
		if amt != 0 {
			t.Fatal("Expected 0, got", amt)
		}
	}

	// Test saving and loading a reading.
	err = c.saveReading(5, 500)
	if err != nil {
		t.Fatal(err)
	}

	// Verify the saved reading.
	amt, err := c.loadReading(5)
	if err != nil {
		t.Fatal(err)
	}
	if amt != 500 {
		t.Fatal("Expected 500, got", amt)
	}

	// Test saving the same reading twice.
	err = c.saveReading(5, 500)
	if err != nil {
		t.Fatal("Saving the same reading twice should not result in error")
	}

	// Test saving a different reading in the same timeslot.
	err = c.saveReading(5, 400)
	if err == nil {
		t.Fatal("Expected error when saving a different reading in the same timeslot")
	}

	// Test reading from an uninitialized timeslot.
	amt, err = c.loadReading(99)
	if err != nil || amt != 0 {
		t.Fatal("Expected 0, got", amt, "with error", err)
	}

	// Saving and loading multiple readings.
	for i := uint32(10); i < 15; i++ {
		err = c.saveReading(i, i*100)
		if err != nil {
			t.Fatal(err)
		}

		amt, err = c.loadReading(i)
		if err != nil || amt != i*100 {
			t.Fatal("Expected", i*100, "got", amt, "with error", err)
		}
	}

	// Final check: Verify that all non-tested entries are still empty.
	for i := uint32(0); i < 100; i++ {
		if i >= 10 && i < 15 || i == 5 {
			continue
		}
		amt, err := c.loadReading(i)
		if err != nil {
			t.Fatal(err)
		}
		if amt != 0 {
			t.Fatal("Expected 0, got", amt)
		}
	}
}
