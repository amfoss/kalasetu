package graph

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

import "kalasetu/services"

type Resolver struct {
	eventService       services.EventService
	userService        services.UserService
	applicationService services.ApplicationService
	listingService     services.ListingService
	cartService        services.CartService
}

func NewResolver(eventService services.EventService, userService services.UserService, applicationService services.ApplicationService, listingService services.ListingService, cartService services.CartService) *Resolver {
	return &Resolver{
		eventService:       eventService,
		userService:        userService,
		applicationService: applicationService,
		listingService:     listingService,
		cartService:        cartService,
	}
}
