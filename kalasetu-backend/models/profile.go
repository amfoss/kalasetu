package models

import "time"

type Achievement struct {
	Title       string
	Description string
	IconType    string
}

type ProfilePost struct {
	ID           int
	Content      string
	MediaType    string
	MediaURI     string
	LikeCount    int
	CommentCount int
	CreatedAt    time.Time
}

type Profile struct {
	ID             int
	Name           string
	UserName       string
	Location       string
	Bio            string
	ProfilePicture string
	Email          string
	Followers      int
	Following      int
	ArtworksCount  int
	ArtworksImages []string
	TotalLikes     int
	Skills         []string
	RecentPosts    []ProfilePost
	Achievements   []Achievement
}