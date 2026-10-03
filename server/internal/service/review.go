package service

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	model "shijibu/internal/model/pgsql"
	"shijibu/internal/platform/wechat"

	"gorm.io/gorm"
)

type TextSafetyProvider interface {
	CheckText(context.Context, string, string) (wechat.TextSafetyResult, error)
}

type ReviewService struct {
	db     *gorm.DB
	safety TextSafetyProvider
}

type ReviewView struct {
	Review  model.RestaurantReview        `json:"review"`
	Version model.RestaurantReviewVersion `json:"version"`
	Place   model.Place                   `json:"place"`
	Tags    []model.Tag                   `json:"tags"`
	Media   []model.RecordMedia           `json:"media"`
}

func NewReviewService(db *gorm.DB, safety TextSafetyProvider) *ReviewService {
	return &ReviewService{db: db, safety: safety}
}

// CreateFromVisit 从一次私人足迹创建公开评论申请；来源足迹保持私有且不发生变化，
// 评论内容和审核状态在同一事务中写入独立评论表。
func (s *ReviewService) CreateFromVisit(ctx context.Context, userID, visitID uint) (*ReviewView, error) {
	var existing model.RestaurantReview
	if err := s.db.Where("visit_record_id = ? AND user_id = ?", visitID, userID).First(&existing).Error; err == nil {
		return s.loadView(existing)
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	visit, source, user, err := s.loadReviewSource(userID, visitID)
	if err != nil {
		return nil, err
	}
	provider, result, rawReference, riskLabels, err := s.checkReviewText(ctx, user, source.Content)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	review := model.RestaurantReview{
		ID: visit.ID, UserID: userID, PlaceID: visit.PlaceID, VisitRecordID: visit.ID,
		PublishStatus: "pending", RiskLevel: riskLevel(riskLabels), SubmittedAt: now,
	}
	version := model.RestaurantReviewVersion{
		ID: source.ID, ReviewID: review.ID, VersionNo: 1, EditorUserID: userID,
		VisitDate: source.VisitDate, Conclusion: source.Conclusion, AverageCost: source.AverageCost,
		WaitMinutes: source.WaitMinutes, MealPeriod: source.MealPeriod, Dishes: source.Dishes,
		Content: source.Content, ChangeSummary: source.ChangeSummary, CreatedAt: now,
	}
	if err := s.saveReview(review, version, source.ID, provider, result, rawReference, riskLabels, now); err != nil {
		return nil, err
	}
	review.CurrentVersionID = &version.ID
	return s.loadView(review)
}

func (s *ReviewService) loadReviewSource(userID, visitID uint) (model.VisitRecord, model.VisitRecordVersion, model.User, error) {
	var visit model.VisitRecord
	if err := s.db.Where("id = ? AND user_id = ?", visitID, userID).First(&visit).Error; err != nil {
		return visit, model.VisitRecordVersion{}, model.User{}, mapNotFound(err)
	}
	if visit.CurrentVersionID == nil {
		return visit, model.VisitRecordVersion{}, model.User{}, ErrInvalidInput
	}
	var source model.VisitRecordVersion
	if err := s.db.Where("id = ? AND record_id = ?", *visit.CurrentVersionID, visit.ID).First(&source).Error; err != nil {
		return visit, source, model.User{}, mapNotFound(err)
	}
	var user model.User
	if err := s.db.First(&user, userID).Error; err != nil || user.AccountStatus != "active" {
		return visit, source, user, ErrForbidden
	}
	return visit, source, user, nil
}

func (s *ReviewService) checkReviewText(ctx context.Context, user model.User, content string) (string, string, string, []string, error) {
	labels := recordRiskLabels(content)
	if strings.TrimSpace(content) == "" {
		return "local_rules", "not_applicable", "", labels, nil
	}
	if s.safety == nil {
		return "local_rules", "passed", "", labels, nil
	}
	result, err := s.safety.CheckText(ctx, user.OpenID, content)
	if err != nil {
		return "", "", "", nil, ErrUnavailable
	}
	if result.Suggest != "pass" {
		labels = append(labels, "wechat_"+strconv.Itoa(result.RiskLabel))
	}
	return "wechat", result.Suggest, strconv.Itoa(result.RiskLabel), labels, nil
}

// saveReview 原子保存不可变评论快照、复制的媒体关系、内容安全结果和审核任务。
func (s *ReviewService) saveReview(review model.RestaurantReview, version model.RestaurantReviewVersion, sourceVersionID uint, provider, result, rawReference string, riskLabels []string, now time.Time) error {
	riskJSON, _ := json.Marshal(riskLabels)
	return s.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&review).Error; err != nil {
			return err
		}
		if err := tx.Create(&version).Error; err != nil {
			return err
		}
		review.CurrentVersionID = &version.ID
		if err := tx.Model(&review).Update("current_version_id", version.ID).Error; err != nil {
			return err
		}
		if err := copyReviewTags(tx, sourceVersionID, version.ID); err != nil {
			return err
		}
		if err := copyReviewMedia(tx, review.VisitRecordID, review.ID); err != nil {
			return err
		}
		check := model.ContentSafetyCheck{TargetType: "restaurant_review_version", TargetID: version.ID, Provider: provider, CheckType: "text", Result: result, RiskLabels: model.JSONDocument(riskJSON), RawResponseRef: rawReference, CheckedAt: now}
		if len(riskLabels) > 0 {
			check.Result = "review"
		}
		if err := tx.Create(&check).Error; err != nil {
			return err
		}
		dueAt := now.Add(24 * time.Hour)
		return tx.Create(&model.ModerationTask{
			TaskType: "review_publish", TargetType: "restaurant_review", TargetID: review.ID,
			ReviewVersionID: &version.ID, Priority: len(riskLabels) * 10,
			RiskLabels: model.JSONDocument(riskJSON), Status: "pending", DueAt: &dueAt,
		}).Error
	})
}

func (s *ReviewService) ListMine(userID uint, status string, limit int) ([]ReviewView, error) {
	if limit < 1 || limit > 100 {
		limit = 50
	}
	db := s.db.Where("user_id = ?", userID)
	if status = strings.TrimSpace(status); status != "" {
		db = db.Where("status = ?", status)
	}
	var reviews []model.RestaurantReview
	if err := db.Order("id DESC").Limit(limit).Find(&reviews).Error; err != nil {
		return nil, err
	}
	views := make([]ReviewView, 0, len(reviews))
	for _, review := range reviews {
		view, err := s.loadView(review)
		if err != nil {
			return nil, err
		}
		views = append(views, *view)
	}
	return views, nil
}

func (s *ReviewService) GetOwned(userID, reviewID uint) (*ReviewView, error) {
	var review model.RestaurantReview
	if err := s.db.Where("id = ? AND user_id = ?", reviewID, userID).First(&review).Error; err != nil {
		return nil, mapNotFound(err)
	}
	return s.loadView(review)
}

func (s *ReviewService) loadView(review model.RestaurantReview) (*ReviewView, error) {
	if review.CurrentVersionID == nil {
		return nil, ErrNotFound
	}
	view := &ReviewView{Review: review}
	if err := s.db.First(&view.Version, *review.CurrentVersionID).Error; err != nil {
		return nil, err
	}
	if err := s.db.First(&view.Place, review.PlaceID).Error; err != nil {
		return nil, err
	}
	if err := s.db.Table("tags").Joins("JOIN restaurant_review_tag_links links ON links.tag_id = tags.id").Where("links.review_version_id = ?", view.Version.ID).Order("tags.sort_order, tags.id").Find(&view.Tags).Error; err != nil {
		return nil, err
	}
	if err := s.db.Table("record_media AS media").Joins("JOIN restaurant_review_media links ON links.media_id = media.id").Where("links.review_id = ?", review.ID).Order("media.sort_order, media.id").Find(&view.Media).Error; err != nil {
		return nil, err
	}
	return view, nil
}

func copyReviewTags(tx *gorm.DB, sourceVersionID, reviewVersionID uint) error {
	var links []model.VisitRecordTagLink
	if err := tx.Where("record_version_id = ?", sourceVersionID).Find(&links).Error; err != nil {
		return err
	}
	for _, link := range links {
		if err := tx.Create(&model.RestaurantReviewTagLink{ReviewVersionID: reviewVersionID, TagID: link.TagID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func copyReviewMedia(tx *gorm.DB, visitID, reviewID uint) error {
	var media []model.RecordMedia
	if err := tx.Where("record_id = ?", visitID).Find(&media).Error; err != nil {
		return err
	}
	for _, item := range media {
		if err := tx.Create(&model.RestaurantReviewMedia{ReviewID: reviewID, MediaID: item.ID}).Error; err != nil {
			return err
		}
	}
	return nil
}

func riskLevel(labels []string) string {
	if len(labels) == 0 {
		return "low"
	}
	if len(labels) == 1 {
		return "medium"
	}
	return "high"
}

func recordRiskLabels(content string) []string {
	content = strings.ToLower(content)
	rules := map[string][]string{
		"serious_accusation": {"中毒", "诈骗", "违法", "违禁品", "黑店"},
		"personal_attack":    {"骗子", "垃圾", "人渣", "去死"},
		"privacy":            {"手机号", "电话", "微信号", "身份证"},
	}
	labels := make([]string, 0)
	for label, words := range rules {
		for _, word := range words {
			if strings.Contains(content, word) {
				labels = append(labels, label)
				break
			}
		}
	}
	return labels
}
