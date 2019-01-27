// +build integration

package main

import (
	"natschat/test"
	"testing"
)

func TestEnvironmentSetup(t *testing.T) {
	db := test.GetTestDB()

	row := db.DB().QueryRow("select 1")
	var i int
	err := row.Scan(&i)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	if i != 1 {
		t.Fatal("failed to scan row")
	}
}

func TestEnvironmentSetupAgain(t *testing.T) {
	db := test.GetTestDB()

	row := db.DB().QueryRow("select 1")
	var i int
	err := row.Scan(&i)
	if err != nil {
		t.Fatalf("failed to get db: %v", err)
	}

	if i != 1 {
		t.Fatal("failed to scan row")
	}
}
