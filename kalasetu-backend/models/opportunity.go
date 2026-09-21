package models

import "time"

type Opportunity struct {
	ID                int       `json:"id"`
	EventID           int       `json:"event_id"`
	Title             string    `json:"title"`
	Description       string    `json:"description"`
	Categories        string    `json:"categories"` // comma separated
	Location          string    `json:"location"`
	StartDate         string    `json:"start_date"`
	EndDate           string    `json:"end_date"`
	TotalPositions    int       `json:"total_positions"`
	OpenSlots         int       `json:"open_slots"`
	ApplicationsCount int       `json:"applications_count"`
	Status            string    `json:"status"`
	HostID            int       `json:"host_id"`
	CreatedAt         time.Time `json:"created_at"`
}

type CreateOpportunityInput struct {
	EventID        int
	Title          string
	Description    string
	Categories     []string
	Location       string
	StartDate      string
	EndDate        string
	TotalPositions int
	Status         string
}

type UpdateOpportunityInput struct {
	Title          *string
	Description    *string
	Categories     []string
	Location       *string
	StartDate      *string
	EndDate        *string
	TotalPositions *int
	Status         *string
}
