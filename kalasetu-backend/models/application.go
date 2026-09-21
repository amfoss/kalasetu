package models

import "time"

type Application struct {
	ID             int       `json:"id"`
	OpportunityID  *int      `json:"opportunity_id,omitempty"`
	EventID        int       `json:"event_id"`
	ApplierID      int       `json:"applier_id"`
	ApplicantName  string    `json:"applicant_name"`
	ApplicantEmail string    `json:"applicant_email"`
	ApplicantPhone string    `json:"applicant_phone"`
	Description    string    `json:"description"`
	ResumeURL      string    `json:"resume_url"`
	Status         string    `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
}

type CreateApplicationInput struct {
	OpportunityID  *int
	EventID        int
	ApplicantName  string
	ApplicantEmail string
	ApplicantPhone string
	Description    string
	ResumeURL      string
}
