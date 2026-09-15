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

	"github.com/99designs/gqlgen/graphql"
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
		Achievements:   []*model.Achievement{},
		RecentPosts:    toGraphProfilePosts(p.RecentPosts),
	}
}
