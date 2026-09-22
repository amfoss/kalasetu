package app_test

import (
	"testing"

	"kalasetu/testutil"
)

const createEventMutation = `
mutation($input: CreateEventInput!) {
  createEvent(input: $input) { id name }
}`

const eventQuery = `
query($id: ID!) {
  event(id: $id) { id name startDate duration bannerUrl }
}`

// Every event read selects the banner columns, so reading an event back is
// enough to prove the banner migration applied: a missing column fails the
// query outright, without needing object storage or an upload.
func TestOrganizerCreatesEventAndReadsItBack(t *testing.T) {
	h := testutil.NewHarness(t)
	organizer := h.CreateUser(t, "Event Organizer")

	var created struct {
		CreateEvent struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"createEvent"`
	}
	h.GraphQL(t, organizer.Token, createEventMutation, map[string]any{
		"input": map[string]any{
			"name":      "Kathakali Evening",
			"startDate": "2026-11-02",
			"duration":  "2 days",
		},
	}, &created)

	if created.CreateEvent.ID == "" {
		t.Fatal("createEvent returned no id")
	}

	var read struct {
		Event *struct {
			ID        string  `json:"id"`
			Name      string  `json:"name"`
			StartDate string  `json:"startDate"`
			Duration  string  `json:"duration"`
			BannerURL *string `json:"bannerUrl"`
		} `json:"event"`
	}
	h.GraphQL(t, organizer.Token, eventQuery, map[string]any{"id": created.CreateEvent.ID}, &read)

	if read.Event == nil {
		t.Fatalf("event %s not found after creating it", created.CreateEvent.ID)
	}
	if read.Event.Name != "Kathakali Evening" {
		t.Fatalf("name = %q, want Kathakali Evening", read.Event.Name)
	}
	if read.Event.StartDate != "2026-11-02" {
		t.Fatalf("startDate = %q, want 2026-11-02", read.Event.StartDate)
	}
}

// A missing image must never block publishing, so the banner is optional.
func TestOrganizerCreatesEventWithoutBanner(t *testing.T) {
	h := testutil.NewHarness(t)
	organizer := h.CreateUser(t, "Event Organizer")

	var created struct {
		CreateEvent struct {
			ID string `json:"id"`
		} `json:"createEvent"`
	}
	h.GraphQL(t, organizer.Token, createEventMutation, map[string]any{
		"input": map[string]any{
			"name":      "Unbannered Recital",
			"startDate": "2026-12-01",
			"duration":  "1 day",
		},
	}, &created)

	var read struct {
		Event *struct {
			BannerURL *string `json:"bannerUrl"`
		} `json:"event"`
	}
	h.GraphQL(t, organizer.Token, eventQuery, map[string]any{"id": created.CreateEvent.ID}, &read)

	if read.Event == nil {
		t.Fatal("event not found after creating it without a banner")
	}
	if read.Event.BannerURL != nil {
		t.Fatalf("bannerUrl = %q, want null for an event created with no banner", *read.Event.BannerURL)
	}
}
