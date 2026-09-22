package graph

import (
	"context"
	"errors"
	"fmt"
	"kalasetu/graph/model"
	"kalasetu/middlewares"
	"kalasetu/models"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/graphql"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

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

func parsePostID(id string) (int, error) {
	parsed, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid post id: %s", id)
	}
	return parsed, nil
}

func parseCommentID(id string) (int, error) {
	parsed, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid comment id: %s", id)
	}
	return parsed, nil
}

func parseProfileID(id string) (int, error) {
	parsed, err := strconv.Atoi(id)
	if err != nil {
		return 0, fmt.Errorf("invalid profile id: %s", id)
	}
	return parsed, nil
}

func parseOptionalID(id *string) (*int, error) {
	if id == nil {
		return nil, nil
	}
	parsed, err := strconv.Atoi(*id)
	if err != nil {
		return nil, fmt.Errorf("invalid id: %s", *id)
	}
	return &parsed, nil
}

func int32PtrToIntPtr(i *int32) *int {
	if i == nil {
		return nil
	}
	v := int(*i)
	return &v
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
	var bannerURL *string
	if e.BannerURL != "" {
		bannerURL = &e.BannerURL
	}
	return &model.Event{
		ID:        strconv.Itoa(e.ID),
		Name:      e.Name,
		StartDate: e.StartDate,
		Duration:  e.Duration,
		HostID:    hostID,
		HostName:  hostName,
		BannerURL: bannerURL,
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
	var oppID *string
	if app.OpportunityID != nil && *app.OpportunityID > 0 {
		str := strconv.Itoa(*app.OpportunityID)
		oppID = &str
	}
	return &model.Application{
		ID:             strconv.Itoa(app.ID),
		OpportunityID:  oppID,
		EventID:        strconv.Itoa(app.EventID),
		ApplierID:      strconv.Itoa(app.ApplierID),
		ApplicantName:  &app.ApplicantName,
		ApplicantEmail: &app.ApplicantEmail,
		ApplicantPhone: &app.ApplicantPhone,
		Description:    &app.Description,
		ResumeURL:      &app.ResumeURL,
		Status:         app.Status,
		CreatedAt:      app.CreatedAt.Format(time.RFC3339),
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
		Archived:  l.Archived,
	}
}

func toGraphListings(listings []models.Listing) []*model.Listing {
	result := make([]*model.Listing, 0, len(listings))
	for i := range listings {
		result = append(result, toGraphListing(&listings[i]))
	}
	return result
}

func toGraphOpportunity(opp *models.Opportunity) *model.Opportunity {
	if opp == nil {
		return nil
	}
	cats := make([]string, 0)
	if opp.Categories != "" {
		cats = strings.Split(opp.Categories, ",")
	}
	return &model.Opportunity{
		ID:                strconv.Itoa(opp.ID),
		EventID:           strconv.Itoa(opp.EventID),
		Title:             opp.Title,
		Description:       &opp.Description,
		Categories:        cats,
		Location:          &opp.Location,
		StartDate:         &opp.StartDate,
		EndDate:           &opp.EndDate,
		TotalPositions:    int32(opp.TotalPositions),
		OpenSlots:         int32(opp.OpenSlots),
		ApplicationsCount: int32(opp.ApplicationsCount),
		Status:            opp.Status,
		CreatedAt:         opp.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphOpportunities(opps []models.Opportunity) []*model.Opportunity {
	result := make([]*model.Opportunity, 0, len(opps))
	for i := range opps {
		o := opps[i]
		result = append(result, toGraphOpportunity(&o))
	}
	return result
}

func toGraphCart(c *models.Cart) *model.Cart {
	lines := make([]*model.CartLine, 0, len(c.Lines))
	for i := range c.Lines {
		l := &c.Lines[i]
		line := &model.CartLine{Listing: toGraphListing(&l.Listing), Quantity: int32(l.Quantity)}
		if l.Issue != "" {
			issue := model.CartLineIssue(l.Issue)
			line.Issue = &issue
		}
		lines = append(lines, line)
	}
	return &model.Cart{Lines: lines, Total: c.Total}
}

func toGraphShipping(s models.ShippingAddress) *model.ShippingAddress {
	return &model.ShippingAddress{
		Name: s.Name, Phone: s.Phone, Line1: s.Line1, Line2: s.Line2, City: s.City,
		State: s.State, PostalCode: s.PostalCode, Country: s.Country,
	}
}

func toGraphSellerOrderItem(it *models.SellerOrderItem) *model.SellerOrderItem {
	return &model.SellerOrderItem{
		ID:              strconv.Itoa(it.ID),
		OrderID:         strconv.Itoa(it.OrderID),
		ListingID:       strconv.Itoa(it.ListingID),
		Title:           it.Title,
		Price:           it.Price,
		Quantity:        int32(it.Quantity),
		Status:          model.FulfilmentStatus(it.Status),
		ShippingAddress: toGraphShipping(it.Shipping),
		CreatedAt:       it.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphOrder(o *models.Order) *model.Order {
	items := make([]*model.OrderItem, 0, len(o.Items))
	for _, it := range o.Items {
		items = append(items, &model.OrderItem{
			ID:        strconv.Itoa(it.ID),
			ListingID: strconv.Itoa(it.ListingID),
			Title:     it.Title,
			Price:     it.Price,
			Quantity:  int32(it.Quantity),
			Status:    model.FulfilmentStatus(it.Status),
		})
	}
	s := o.Shipping
	return &model.Order{
		ID:              strconv.Itoa(o.ID),
		Items:           items,
		ShippingAddress: toGraphShipping(s),
		Total:           o.Total,
		State:           model.OrderState(o.State()),
		CreatedAt:       o.CreatedAt.Format(time.RFC3339),
	}
}

// optionalUserID returns the authenticated user id from context, or 0 if not authenticated.
func optionalUserID(ctx context.Context) int {
	userID, err := middlewares.GetUserIDFromContext(ctx)
	if err != nil {
		return 0
	}
	return userID
}

func toGraphPost(p *models.Post) *model.Post {
	if p == nil {
		return nil
	}
	var categoryID *string
	if p.CategoryID != nil {
		id := strconv.Itoa(*p.CategoryID)
		categoryID = &id
	}
	var categoryName *string
	if p.CategoryName != "" {
		categoryName = &p.CategoryName
	}
	media := make([]*model.PostMedia, 0, len(p.Media))
	for i := range p.Media {
		media = append(media, toGraphPostMedia(&p.Media[i]))
	}
	return &model.Post{
		ID:           strconv.Itoa(p.ID),
		UserID:       strconv.Itoa(p.UserID),
		UserName:     p.UserName,
		Content:      p.Content,
		Media:        media,
		CategoryID:   categoryID,
		CategoryName: categoryName,
		LikeCount:    int32(p.LikeCount),
		CommentCount: int32(p.CommentCount),
		IsLikedByMe:  p.IsLikedByMe,
		CreatedAt:    p.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphPostMedia(m *models.PostMedia) *model.PostMedia {
	if m == nil {
		return nil
	}
	return &model.PostMedia{
		ID:        strconv.Itoa(m.ID),
		PostID:    strconv.Itoa(m.PostID),
		URL:       m.URL,
		MediaType: m.MediaType,
		SortOrder: int32(m.SortOrder),
		CreatedAt: m.CreatedAt.Format(time.RFC3339),
	}
}

// toUploadMedia converts a gqlgen Upload value into the application-level
// UploadMedia representation so services stay decoupled from gqlgen. A nil
// upload (no file attached) yields nil.
func toUploadMedia(u *graphql.Upload) *models.UploadMedia {
	if u == nil {
		return nil
	}
	return &models.UploadMedia{
		Reader:      u.File,
		Filename:    u.Filename,
		ContentType: u.ContentType,
	}
}

// toUploadMediaList converts gqlgen's Upload values into the application-level
// UploadMedia representation so services stay decoupled from gqlgen.
func toUploadMediaList(uploads []*graphql.Upload) []models.UploadMedia {
	if len(uploads) == 0 {
		return []models.UploadMedia{}
	}
	result := make([]models.UploadMedia, 0, len(uploads))
	for _, u := range uploads {
		if m := toUploadMedia(u); m != nil {
			result = append(result, *m)
		}
	}
	return result
}

func toGraphPosts(posts []models.Post) []*model.Post {
	result := make([]*model.Post, 0, len(posts))
	for i := range posts {
		p := posts[i]
		result = append(result, toGraphPost(&p))
	}
	return result
}

func toGraphAuthor(a *models.Author) *model.Author {
	if a == nil {
		return nil
	}
	return &model.Author{
		ID:   strconv.Itoa(a.ID),
		Name: a.Name,
	}
}

func toGraphAuthors(authors []models.Author) []*model.Author {
	result := make([]*model.Author, 0, len(authors))
	for i := range authors {
		a := authors[i]
		result = append(result, toGraphAuthor(&a))
	}
	return result
}

func toGraphComment(c *models.Comment) *model.Comment {
	if c == nil {
		return nil
	}
	return &model.Comment{
		ID:        strconv.Itoa(c.ID),
		PostID:    strconv.Itoa(c.PostID),
		UserID:    strconv.Itoa(c.UserID),
		UserName:  c.UserName,
		Content:   c.Content,
		CreatedAt: c.CreatedAt.Format(time.RFC3339),
	}
}

func toGraphComments(comments []models.Comment) []*model.Comment {
	result := make([]*model.Comment, 0, len(comments))
	for i := range comments {
		c := comments[i]
		result = append(result, toGraphComment(&c))
	}
	return result
}

func toGraphProfilePosts(posts []models.ProfilePost) []*model.ProfilePost {
	result := make([]*model.ProfilePost, 0, len(posts))
	for i := range posts {
		p := posts[i]
		post := &model.ProfilePost{
			ID:           strconv.Itoa(p.ID),
			Content:      p.Content,
			LikeCount:    int32(p.LikeCount),
			CommentCount: int32(p.CommentCount),
			CreatedAt:    p.CreatedAt.Format(time.RFC3339),
		}
		if p.MediaType != "" {
			post.MediaType = &p.MediaType
		}
		if p.MediaURI != "" {
			post.MediaURI = &p.MediaURI
		}
		result = append(result, post)
	}
	return result
}

func toGraphAchievements(items []models.Achievement) []*model.Achievement {
	result := make([]*model.Achievement, 0, len(items))
	for _, a := range items {
		result = append(result, &model.Achievement{
			Title:       a.Title,
			Description: a.Description,
			IconType:    model.AchievementIcon(a.IconType),
		})
	}
	return result
}

func toGraphProfile(p *models.Profile) *model.Profile {
	var avatarURL *string
	if p.ProfilePicture != "" {
		avatarURL = &p.ProfilePicture
	}
	return &model.Profile{
		ID:             strconv.Itoa(p.ID),
		Name:           p.Name,
		Username:       p.UserName,
		Location:       p.Location,
		Bio:            p.Bio,
		AvatarURL:      avatarURL,
		Email:          p.Email,
		Followers:      int32(p.Followers),
		Following:      int32(p.Following),
		ArtworksCount:  int32(p.ArtworksCount),
		TotalLikes:     int32(p.TotalLikes),
		Skills:         p.Skills,
		ArtworksImages: p.ArtworksImages,
		Achievements:   toGraphAchievements(p.Achievements),
		RecentPosts:    toGraphProfilePosts(p.RecentPosts),
	}
}
