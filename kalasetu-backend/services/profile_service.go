package services

import (
	"context"
	"errors"
	"strings"
	"kalasetu/models"
	"kalasetu/repos"
	"kalasetu/storage"
)

var ErrProfileNotFound = errors.New("profile not found")

const defaultRecentPostsLimit = 5

type ProfileService interface {
	GetProfile(ctx context.Context, userID int) (*models.Profile, error)
}

type profileService struct {
	profileRepo repos.ProfileRepository
	storage     storage.ObjectStorage
}

func NewProfileService(profileRepo repos.ProfileRepository, objectStorage storage.ObjectStorage) ProfileService {
	return &profileService{
		profileRepo: profileRepo,
		storage:     objectStorage,
	}
}

func (s *profileService) GetProfile(ctx context.Context, userID int) (*models.Profile, error) {
	profile, err := s.profileRepo.GetBasics(ctx, userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, ErrProfileNotFound
	}

	profile.ArtworksCount, err = s.profileRepo.CountArtworks(ctx, userID)
	if err != nil {
		return nil, err
	}

	artworksKeys, err := s.profileRepo.ListArtworkImages(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.ArtworksImages = make([]string, 0, len(artworksKeys))
	for _, key := range artworksKeys {
		profile.ArtworksImages = append(profile.ArtworksImages, mediaURL(ctx, s.storage, key))
	}

	profile.TotalLikes, err = s.profileRepo.CountTotalLikes(ctx, userID)
	if err != nil {
		return nil, err
	}

	recentPosts, err := s.profileRepo.ListRecentPosts(ctx, userID, defaultRecentPostsLimit)
	if err != nil {
		return nil, err
	}

	for i := range recentPosts {
		recentPosts[i].MediaURI = mediaURL(ctx, s.storage, recentPosts[i].MediaURI)
	}
	profile.RecentPosts = recentPosts

	profile.Followers, err = s.profileRepo.CountFollowers(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.Following, err = s.profileRepo.CountFollowing(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.Skills, err = s.profileRepo.ListSkills(ctx, userID)
	if err != nil {
		return nil, err
	}

	profile.Achievements, err = s.profileRepo.ListAchievements(ctx, userID)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

// mediaURL resolves a stored object key to a public URL. Keys that are already
// absolute URLs (e.g. seeded placeholder images) are returned unchanged.
func mediaURL(ctx context.Context, s storage.ObjectStorage, key string) string {
	if key == "" || s == nil {
		return key
	}
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return key
	}
	if url, err := s.GetURL(ctx, key); err == nil {
		return url
	}
	return key
}