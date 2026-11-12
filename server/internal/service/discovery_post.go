package service

import (
	"errors"
	"strconv"
	"strings"

	model "shijibu/internal/model/pgsql"

	"gorm.io/gorm"
)

type DiscoveryPostView struct {
	Post   model.DiscoveryPost        `json:"post"`
	Place  model.Place                `json:"place"`
	Images []model.DiscoveryPostImage `json:"images"`
}

func (s *DiscoveryService) ListDiscoveryPosts(cityCode, cursor string, limit int, placeID uint) (Page[DiscoveryPostView], error) {
	limit = normalizeLimit(limit)
	db := s.db.Model(&model.DiscoveryPost{}).Where("discovery_posts.status = ?", "published")
	if cityCode = strings.TrimSpace(cityCode); cityCode != "" {
		db = db.Joins("JOIN places ON places.id = discovery_posts.place_id").Where("places.city_code = ? AND places.status = ?", cityCode, "active")
	}
	if placeID > 0 {
		db = db.Where("discovery_posts.place_id = ?", placeID)
	}
	if cursorID := parseCursor(cursor); cursorID > 0 {
		db = db.Where("discovery_posts.id < ?", cursorID)
	}
	var posts []model.DiscoveryPost
	if err := db.Order("discovery_posts.id DESC").Limit(limit + 1).Find(&posts).Error; err != nil {
		return Page[DiscoveryPostView]{}, err
	}
	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
	}
	views, err := s.loadDiscoveryPostViews(posts)
	if err != nil {
		return Page[DiscoveryPostView]{}, err
	}
	page := Page[DiscoveryPostView]{Items: views, HasMore: hasMore}
	if hasMore && len(posts) > 0 {
		page.NextCursor = strconv.FormatUint(uint64(posts[len(posts)-1].ID), 10)
	}
	return page, nil
}

func (s *DiscoveryService) GetDiscoveryPost(id uint) (*DiscoveryPostView, error) {
	if id == 0 {
		return nil, ErrInvalidInput
	}
	var post model.DiscoveryPost
	if err := s.db.Where("id = ? AND status = ?", id, "published").First(&post).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	views, err := s.loadDiscoveryPostViews([]model.DiscoveryPost{post})
	if err != nil {
		return nil, err
	}
	return &views[0], nil
}

func (s *DiscoveryService) loadDiscoveryPostViews(posts []model.DiscoveryPost) ([]DiscoveryPostView, error) {
	if len(posts) == 0 {
		return []DiscoveryPostView{}, nil
	}
	postIDs := make([]uint, 0, len(posts))
	placeIDs := make([]uint, 0, len(posts))
	for _, post := range posts {
		postIDs = append(postIDs, post.ID)
		placeIDs = append(placeIDs, post.PlaceID)
	}
	var places []model.Place
	if err := s.db.Where("id IN ? AND status = ?", placeIDs, "active").Find(&places).Error; err != nil {
		return nil, err
	}
	placeByID := make(map[uint]model.Place, len(places))
	for _, place := range places {
		placeByID[place.ID] = place
	}
	var images []model.DiscoveryPostImage
	if err := s.db.Where("post_id IN ?", postIDs).Order("post_id, sort_order, id").Find(&images).Error; err != nil {
		return nil, err
	}
	imagesByPost := make(map[uint][]model.DiscoveryPostImage, len(posts))
	for _, image := range images {
		if len(imagesByPost[image.PostID]) < 6 {
			imagesByPost[image.PostID] = append(imagesByPost[image.PostID], image)
		}
	}
	views := make([]DiscoveryPostView, 0, len(posts))
	for _, post := range posts {
		place, ok := placeByID[post.PlaceID]
		if !ok {
			continue
		}
		views = append(views, DiscoveryPostView{Post: post, Place: place, Images: imagesByPost[post.ID]})
	}
	return views, nil
}
