package graph

import (
	"context"
	"errors"
	"fmt"
	"kalasetu/graph/model"
	"kalasetu/middlewares"
	"kalasetu/models"
	"strconv"
	"time"
)

// requireUser returns the authenticated user id from the request context,
// injected by the OptionalJWT middleware, or a GraphQL auth error.
func requireUser(ctx context.Context) (int, error) {
	userID, err := middlewares.GetUserIDFromContext(ctx)
	if err != nil {
		return 0, errors.New("authentication required")
	}
	return userID, nil
}

func parseEventID(id string) (int, error) {
	parsed, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid event id: %s", id)
	}
	return parsed, nil
}

func toGraphEvent(e *models.Event) *model.Event {
	if e == nil {
		return nil
	}
	var hostID *string
	if e.HostID != nil {
		id := strconv.Itoa(*e.HostID)
		hostID = &id
	}
	var hostName *string
	if e.HostName != "" {
		hostName = &e.HostName
	}
	return &model.Event{
		ID:        strconv.Itoa(e.ID),
		Name:      e.Name,
		StartDate: e.StartDate,
		Duration:  e.Duration,
		HostID:    hostID,
		HostName:  hostName,
		CreatedAt: e.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphEvents(events []models.Event) []*model.Event {
	result := make([]*model.Event, 0, len(events))
	for i := range events {
		e := events[i]
		result = append(result, toGraphEvent(&e))
	}
	return result
}

func toGraphApplication(app *models.Application) *model.Application {
	if app == nil {
		return nil
	}
	return &model.Application{
		ID:            strconv.Itoa(app.ID),
		OpportunityID: strconv.Itoa(app.OpportunityID),
		ApplierID:     strconv.Itoa(app.ApplierID),
		ResumeURL:     app.ResumeURL,
		Status:        app.Status,
		CreatedAt:     app.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphApplications(apps []models.Application) []*model.Application {
	result := make([]*model.Application, 0, len(apps))
	for i := range apps {
		a := apps[i]
		result = append(result, toGraphApplication(&a))
	}
	return result
}

func toGraphCategories(categories []models.Category) []*model.Category {
	result := make([]*model.Category, 0, len(categories))
	for _, c := range categories {
		result = append(result, &model.Category{ID: strconv.Itoa(c.ID), Name: c.Name})
	}
	return result
}

func toGraphListing(l *models.Listing) *model.Listing {
	if l == nil {
		return nil
	}
	var location, picture *string
	if l.Seller.Location != "" {
		location = &l.Seller.Location
	}
	if l.Seller.ProfilePicture != "" {
		picture = &l.Seller.ProfilePicture
	}
	return &model.Listing{
		ID:          strconv.Itoa(l.ID),
		Title:       l.Title,
		Description: l.Description,
		Price:       l.Price,
		Currency:    l.Currency,
		Stock:       int32(l.Stock),
		ImageUrls:   l.ImageURLs,
		Category:    &model.Category{ID: strconv.Itoa(l.Category.ID), Name: l.Category.Name},
		Seller: &model.Seller{
			ID:             strconv.Itoa(l.Seller.ID),
			Name:           l.Seller.Name,
			Location:       location,
			ProfilePicture: picture,
		},
		CreatedAt: l.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphListings(listings []models.Listing) []*model.Listing {
	result := make([]*model.Listing, 0, len(listings))
	for i := range listings {
		result = append(result, toGraphListing(&listings[i]))
	}
	return result
}
