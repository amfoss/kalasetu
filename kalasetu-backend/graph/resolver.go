package graph

import "kalasetu/services"

type Resolver struct {
	eventService       services.EventService
	applicationService services.ApplicationService
	opportunityService services.OpportunityService
	postService        services.PostService
	commentService     services.CommentService
	likeService        services.LikeService
	userService        services.UserService
	profileService     services.ProfileService
	listingService     services.ListingService
	cartService        services.CartService
	orderService       services.OrderService
}

func NewResolver(
	eventService services.EventService,
	applicationService services.ApplicationService,
	opportunityService services.OpportunityService,
	postService services.PostService,
	commentService services.CommentService,
	likeService services.LikeService,
	userService services.UserService,
	profileService services.ProfileService,
	listingService services.ListingService,
	cartService services.CartService,
	orderService services.OrderService,
) *Resolver {
	return &Resolver{
		eventService:       eventService,
		applicationService: applicationService,
		opportunityService: opportunityService,
		postService:        postService,
		commentService:     commentService,
		likeService:        likeService,
		userService:        userService,
		profileService:     profileService,
		listingService:     listingService,
		cartService:        cartService,
		orderService:       orderService,
	}
}
