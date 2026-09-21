package services

import (
	"context"
	"errors"
	"fmt"
	"kalasetu/models"
	"kalasetu/repos"
	"kalasetu/storage"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

var (
	ErrPostNotFound        = errors.New("post not found")
	ErrCommentNotFound     = errors.New("comment not found")
	ErrPostForbidden       = errors.New("you are not the owner of this post")
	ErrCommentForbidden    = errors.New("you are not the owner of this comment")
	ErrStorageUnconfigured = errors.New("object storage is not configured")
)

const maxMediaPerPost = 10

type PostService interface {
	Create(ctx context.Context, userID int, input models.CreatePostInput) (*models.Post, error)
	GetByID(ctx context.Context, id int, currentUserID int) (*models.Post, error)
	List(ctx context.Context, currentUserID int, limit, offset *int) ([]models.Post, error)
	ListByUser(ctx context.Context, authorUserID int, currentUserID int, limit, offset *int) ([]models.Post, error)
	Update(ctx context.Context, userID, id int, input models.UpdatePostInput) (*models.Post, error)
	Delete(ctx context.Context, userID, id int) error
}

type postService struct {
	postRepo      repos.PostRepository
	postMediaRepo repos.PostMediaRepository
	storage       storage.ObjectStorage
}

func NewPostService(
	postRepo repos.PostRepository,
	postMediaRepo repos.PostMediaRepository,
	objectStorage storage.ObjectStorage,
) PostService {
	return &postService{
		postRepo:      postRepo,
		postMediaRepo: postMediaRepo,
		storage:       objectStorage,
	}
}

// Create persists the post, uploads each attached file to object storage,
// records the resulting object keys in post_media and returns the post with
// its media. Upload order is preserved via SortOrder.
func (s *postService) Create(ctx context.Context, userID int, input models.CreatePostInput) (*models.Post, error) {
	if len(input.Media) > int(maxMediaPerPost) {
		return nil, fmt.Errorf("a post can have at most %d media files", maxMediaPerPost)
	}

	post, err := s.postRepo.Create(ctx, &models.Post{
		UserID:     userID,
		Content:    input.Content,
		CategoryID: input.CategoryID,
	})
	if err != nil {
		return nil, err
	}

	if err := s.attachMedia(ctx, post.ID, input.Media); err != nil {
		return nil, err
	}

	return s.getPost(ctx, post.ID, userID)
}

// attachMedia uploads each file to object storage and records its object key
// in post_media. If any step fails, already-uploaded objects are removed so no
// orphaned files remain.
func (s *postService) attachMedia(ctx context.Context, postID int, media []models.UploadMedia) error {
	if len(media) == 0 {
		return nil
	}
	if s.storage == nil {
		return ErrStorageUnconfigured
	}

	rows := make([]models.PostMedia, 0, len(media))
	for i, file := range media {
		key, err := s.generateObjectKey(postID, file.Filename)
		if err != nil {
			return err
		}
		if err := s.storage.Upload(ctx, key, file.Reader, file.ContentType); err != nil {
			s.cleanupUploaded(ctx, rows)
			return fmt.Errorf("failed to upload media: %w", err)
		}
		rows = append(rows, models.PostMedia{
			PostID:    postID,
			ObjectKey: key,
			MediaType: file.ContentType,
			SortOrder: i,
		})
	}

	if err := s.postMediaRepo.CreateMany(ctx, postID, rows); err != nil {
		s.cleanupUploaded(ctx, rows)
		return err
	}
	return nil
}

// cleanupUploaded best-effort removes already-uploaded objects when a later
// step in the creation flow fails.
func (s *postService) cleanupUploaded(ctx context.Context, media []models.PostMedia) {
	if s.storage == nil {
		return
	}
	for _, m := range media {
		if err := s.storage.Delete(ctx, m.ObjectKey); err != nil {
			log.Printf("warning: failed to delete orphaned object %q: %v", m.ObjectKey, err)
		}
	}
}

// generateObjectKey builds a unique, storage-friendly key of the form
// posts/{postID}/{uuid}.{ext}. The original filename is never used verbatim.
func (s *postService) generateObjectKey(postID int, filename string) (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", fmt.Errorf("failed to generate object key: %w", err)
	}
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if ext == "" {
		ext = "bin"
	}
	return fmt.Sprintf("posts/%d/%s.%s", postID, id.String(), ext), nil
}

// hydrateMedia loads post_media rows for a post and resolves each object key to
// a public URL.
func (s *postService) hydrateMedia(ctx context.Context, post *models.Post) error {
	if post == nil {
		return nil
	}
	media, err := s.postMediaRepo.ListByPost(ctx, post.ID)
	if err != nil {
		return err
	}
	for i := range media {
		media[i].URL = mediaURL(ctx, s.storage, media[i].ObjectKey)
	}
	post.Media = media
	return nil
}

// getPost fetches a post by id and hydrates its media with public URLs.
func (s *postService) getPost(ctx context.Context, id int, currentUserID int) (*models.Post, error) {
	post, err := s.postRepo.FindByID(ctx, id, currentUserID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	if err := s.hydrateMedia(ctx, post); err != nil {
		return nil, err
	}
	return post, nil
}

func (s *postService) GetByID(ctx context.Context, id int, currentUserID int) (*models.Post, error) {
	return s.getPost(ctx, id, currentUserID)
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func (s *postService) List(ctx context.Context, currentUserID int, limit, offset *int) ([]models.Post, error) {
	posts, err := s.postRepo.List(ctx, currentUserID, derefInt(limit), derefInt(offset))
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return posts, nil
	}

	ids := make([]int, 0, len(posts))
	for i := range posts {
		ids = append(ids, posts[i].ID)
	}

	mediaByPost, err := s.postMediaRepo.ListByPosts(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		media, ok := mediaByPost[posts[i].ID]
		if !ok {
			posts[i].Media = []models.PostMedia{}
			continue
		}
		for j := range media {
			media[j].URL = mediaURL(ctx, s.storage, media[j].ObjectKey)
		}
		posts[i].Media = media
	}
	return posts, nil
}

func (s *postService) ListByUser(ctx context.Context, authorUserID int, currentUserID int, limit, offset *int) ([]models.Post, error) {
	posts, err := s.postRepo.ListByUser(ctx, authorUserID, currentUserID, derefInt(limit), derefInt(offset))
	if err != nil {
		return nil, err
	}
	if len(posts) == 0 {
		return posts, nil
	}

	ids := make([]int, 0, len(posts))
	for i := range posts {
		ids = append(ids, posts[i].ID)
	}

	mediaByPost, err := s.postMediaRepo.ListByPosts(ctx, ids)
	if err != nil {
		return nil, err
	}

	for i := range posts {
		media, ok := mediaByPost[posts[i].ID]
		if !ok {
			posts[i].Media = []models.PostMedia{}
			continue
		}
		for j := range media {
			media[j].URL = mediaURL(ctx, s.storage, media[j].ObjectKey)
		}
		posts[i].Media = media
	}
	return posts, nil
}

func (s *postService) Update(ctx context.Context, userID, id int, input models.UpdatePostInput) (*models.Post, error) {
	post, err := s.postRepo.FindByID(ctx, id, userID)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	if post.UserID != userID {
		return nil, ErrPostForbidden
	}
	if err := s.postRepo.Update(ctx, id, input); err != nil {
		return nil, err
	}
	return s.getPost(ctx, id, userID)
}

func (s *postService) Delete(ctx context.Context, userID, id int) error {
	post, err := s.postRepo.FindByID(ctx, id, userID)
	if err != nil {
		return err
	}
	if post == nil {
		return ErrPostNotFound
	}
	if post.UserID != userID {
		return ErrPostForbidden
	}

	media, err := s.postMediaRepo.ListByPost(ctx, id)
	if err != nil {
		return err
	}

	if err := s.postRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Best-effort removal of the underlying objects; post_media rows are
	// removed by the ON DELETE CASCADE on the post.
	if s.storage != nil {
		s.cleanupUploaded(ctx, media)
	}
	return nil
}

type CommentService interface {
	Create(ctx context.Context, userID int, input models.CreateCommentInput) (*models.Comment, error)
	ListByPost(ctx context.Context, postID int) ([]models.Comment, error)
	Update(ctx context.Context, userID, id int, content string) (*models.Comment, error)
	Delete(ctx context.Context, userID, id int) error
}

type commentService struct {
	commentRepo repos.CommentRepository
}

func NewCommentService(commentRepo repos.CommentRepository) CommentService {
	return &commentService{commentRepo: commentRepo}
}

func (s *commentService) Create(ctx context.Context, userID int, input models.CreateCommentInput) (*models.Comment, error) {
	comment, err := s.commentRepo.Create(ctx, &models.Comment{
		PostID:  input.PostID,
		UserID:  userID,
		Content: input.Content,
	})
	if err != nil {
		return nil, err
	}
	// Refetch so the response includes the commenter name.
	return s.commentRepo.FindByID(ctx, comment.ID)
}

func (s *commentService) ListByPost(ctx context.Context, postID int) ([]models.Comment, error) {
	return s.commentRepo.ListByPost(ctx, postID)
}

func (s *commentService) Update(ctx context.Context, userID, id int, content string) (*models.Comment, error) {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrCommentNotFound
	}
	if comment.UserID != userID {
		return nil, ErrCommentForbidden
	}
	if err := s.commentRepo.Update(ctx, id, content); err != nil {
		return nil, err
	}
	return s.commentRepo.FindByID(ctx, id)
}

func (s *commentService) Delete(ctx context.Context, userID, id int) error {
	comment, err := s.commentRepo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	if comment == nil {
		return ErrCommentNotFound
	}
	if comment.UserID != userID {
		return ErrCommentForbidden
	}
	return s.commentRepo.Delete(ctx, id)
}

type LikeService interface {
	Like(ctx context.Context, userID, postID int) error
	Unlike(ctx context.Context, userID, postID int) error
	ListByPost(ctx context.Context, postID int) ([]models.Author, error)
}

type likeService struct {
	likeRepo repos.LikeRepository
}

func NewLikeService(likeRepo repos.LikeRepository) LikeService {
	return &likeService{likeRepo: likeRepo}
}

func (s *likeService) Like(ctx context.Context, userID, postID int) error {
	return s.likeRepo.Like(ctx, userID, postID)
}

func (s *likeService) Unlike(ctx context.Context, userID, postID int) error {
	return s.likeRepo.Unlike(ctx, userID, postID)
}

func (s *likeService) ListByPost(ctx context.Context, postID int) ([]models.Author, error) {
	return s.likeRepo.ListUsersByPost(ctx, postID)
}
