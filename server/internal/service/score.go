package service

import (
	"fmt"
	"time"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

type ScoreService struct {
	db *gorm.DB
}

func NewScoreService(db *gorm.DB) *ScoreService {
	return &ScoreService{db: db}
}

// DB returns the underlying gorm.DB.
func (s *ScoreService) DB() *gorm.DB {
	return s.db
}

// AddScore awards points to a user and logs the action.
func (s *ScoreService) AddScore(userID uint, action string, points int, desc string) error {
	score := s.getOrCreateScore(userID)
	newTotal := score.TotalScore + points
	newWeek := score.WeekScore + points

	if err := s.db.Model(&score).Updates(map[string]interface{}{
		"total_score": newTotal,
		"week_score":  newWeek,
	}).Error; err != nil {
		return err
	}

	log := model.ScoreLog{
		UserID:      userID,
		Action:      action,
		Points:      points,
		Description: desc,
	}
	return s.db.Create(&log).Error
}

// OnAnalyze is called when a user completes a food analysis.
// Awards: +5 per analysis, +10 for first analysis of the day, streak bonus.
func (s *ScoreService) OnAnalyze(userID uint) error {
	score := s.getOrCreateScore(userID)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	isFirst := score.LastCheckinAt == nil || score.LastCheckinAt.Before(today)

	// Update streak
	streak := score.StreakDays
	if isFirst {
		if score.LastCheckinAt != nil {
			yesterday := today.AddDate(0, 0, -1)
			if score.LastCheckinAt.Format("2006-01-02") == yesterday.Format("2006-01-02") {
				streak++
			} else {
				streak = 1
			}
		} else {
			streak = 1
		}
	}

	// Update score
	updates := map[string]interface{}{
		"streak_days":     streak,
		"last_checkin_at": now,
	}

	if err := s.db.Model(&score).Updates(updates).Error; err != nil {
		return err
	}

	// Award points: +5 per analysis
	if err := s.AddScore(userID, "analyze", 5, "完成食物识别"); err != nil {
		return err
	}

	// First analysis of the day: +10 bonus + streak bonus
	if isFirst {
		if err := s.AddScore(userID, "first_daily", 10, "每日首次识别"); err != nil {
			return err
		}
		if streak > 1 {
			streakBonus := streak * 2
			if streakBonus > 100 {
				streakBonus = 100
			}
			if err := s.AddScore(userID, "streak", streakBonus, fmt.Sprintf("连续打卡%d天奖励", streak)); err != nil {
				return err
			}
		}
	}

	// Check three-meal bonus
	s.checkThreeMeals(userID, today)

	return nil
}

// checkThreeMeals awards +15 if the user has recorded all three meal types today.
func (s *ScoreService) checkThreeMeals(userID uint, today time.Time) {
	tomorrow := today.AddDate(0, 0, 1)
	var mealTypes []int
	s.db.Model(&model.FoodRecord{}).
		Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, today, tomorrow).
		Distinct("meal_type").
		Pluck("meal_type", &mealTypes)

	hasBreakfast, hasLunch, hasDinner := false, false, false
	for _, m := range mealTypes {
		switch m {
		case 1:
			hasBreakfast = true
		case 2:
			hasLunch = true
		case 3:
			hasDinner = true
		}
	}

	if hasBreakfast && hasLunch && hasDinner {
		// Check if already awarded today
		var count int64
		s.db.Model(&model.ScoreLog{}).
			Where("user_id = ? AND action = ? AND created_at >= ? AND created_at < ?", userID, "three_meals", today, tomorrow).
			Count(&count)
		if count == 0 {
			s.AddScore(userID, "three_meals", 15, "三餐全记录奖励")
		}
	}
}

// OnInvite awards points when a new user registers via invitation.
func (s *ScoreService) OnInvite(inviterID, inviteeID uint) error {
	// Award inviter
	if err := s.AddScore(inviterID, "invite", 50, "邀请新用户注册"); err != nil {
		return err
	}

	// Create bidirectional friend relation
	friend := model.FriendRelation{
		UserID:       inviterID,
		FriendUserID: inviteeID,
	}
	s.db.Where("user_id = ? AND friend_user_id = ?", inviterID, inviteeID).FirstOrCreate(&friend)

	friend2 := model.FriendRelation{
		UserID:       inviteeID,
		FriendUserID: inviterID,
	}
	s.db.Where("user_id = ? AND friend_user_id = ?", inviteeID, inviterID).FirstOrCreate(&friend2)

	return nil
}

// OnInviteeAnalyze awards bonus when invited user completes first analysis.
func (s *ScoreService) OnInviteeAnalyze(inviterID, inviteeID uint) error {
	var count int64
	s.db.Model(&model.ScoreLog{}).
		Where("user_id = ? AND action = ? AND description LIKE ?", inviterID, "invitee_analyze", fmt.Sprintf("%%用户%d%%", inviteeID)).
		Count(&count)
	if count > 0 {
		return nil
	}
	return s.AddScore(inviterID, "invitee_analyze", 20, fmt.Sprintf("被邀请用户%d首次识别奖励", inviteeID))
}

// OnShare awards points for sharing a poster.
func (s *ScoreService) OnShare(userID uint) error {
	shareLog := model.ShareRecord{
		UserID:    userID,
		ShareType: "poster",
	}
	if err := s.db.Create(&shareLog).Error; err != nil {
		return err
	}
	return s.AddScore(userID, "share", 5, "分享海报")
}

// GetMyScore returns the current user's score and rank info.
func (s *ScoreService) GetMyScore(userID uint) map[string]interface{} {
	score := s.getOrCreateScore(userID)

	// Calculate global total rank
	var totalRank int64
	s.db.Model(&model.UserScore{}).
		Joins("JOIN users ON users.id = user_scores.user_id").
		Where("users.show_on_rank = true AND total_score > ?", score.TotalScore).
		Count(&totalRank)

	// Calculate global week rank
	var weekRank int64
	s.db.Model(&model.UserScore{}).
		Joins("JOIN users ON users.id = user_scores.user_id").
		Where("users.show_on_rank = true AND week_score > ?", score.WeekScore).
		Count(&weekRank)

	// Get show_on_rank status
	var user model.User
	s.db.Select("show_on_rank").First(&user, userID)

	return map[string]interface{}{
		"total_score":  score.TotalScore,
		"week_score":   score.WeekScore,
		"streak_days":  score.StreakDays,
		"total_rank":   totalRank + 1,
		"week_rank":    weekRank + 1,
		"show_on_rank": user.ShowOnRank,
	}
}

// GetRank returns a paginated rank list.
func (s *ScoreService) GetRank(scope string, offset, limit int) []map[string]interface{} {
	orderCol := "total_score"
	if scope == "week" {
		orderCol = "week_score"
	}

	var scores []model.UserScore
	s.db.Joins("JOIN users ON users.id = user_scores.user_id").
		Where("users.show_on_rank = true").
		Order(orderCol + " DESC").Offset(offset).Limit(limit).Find(&scores)

	var result []map[string]interface{}
	for i, sc := range scores {
		var user model.User
		s.db.Select("id, nickname, avatar_url, title").First(&user, sc.UserID)

		achievementName := ""
		if user.Title > 0 {
			var a model.Achievement
			if err := s.db.First(&a, user.Title).Error; err == nil {
				achievementName = a.Name
			}
		}

		result = append(result, map[string]interface{}{
			"rank":       offset + i + 1,
			"user_id":    sc.UserID,
			"nickname":   user.Nickname,
			"avatar_url": user.AvatarURL,
			"title":      achievementName,
			"score": map[string]interface{}{
				"total": sc.TotalScore,
				"week":  sc.WeekScore,
			}[scope],
			"streak_days": sc.StreakDays,
		})
	}
	return result
}

// GetFriendRank returns rank among friends for a user.
func (s *ScoreService) GetFriendRank(userID uint, offset, limit int) []map[string]interface{} {
	// Get friend IDs
	friendIDs := []uint{userID}
	var friends []model.FriendRelation
	s.db.Where("user_id = ?", userID).Find(&friends)
	for _, f := range friends {
		friendIDs = append(friendIDs, f.FriendUserID)
	}

	var scores []model.UserScore
	s.db.Where("user_id IN ?", friendIDs).
		Joins("JOIN users ON users.id = user_scores.user_id").
		Where("users.show_on_rank = true").
		Order("week_score DESC").
		Offset(offset).Limit(limit).
		Find(&scores)

	var result []map[string]interface{}
	for i, sc := range scores {
		var user model.User
		s.db.Select("id, nickname, avatar_url, title").First(&user, sc.UserID)

		achievementName := ""
		if user.Title > 0 {
			var a model.Achievement
			if err := s.db.First(&a, user.Title).Error; err == nil {
				achievementName = a.Name
			}
		}

		result = append(result, map[string]interface{}{
			"rank":        offset + i + 1,
			"user_id":     sc.UserID,
			"nickname":    user.Nickname,
			"avatar_url":  user.AvatarURL,
			"title":       achievementName,
			"week_score":  sc.WeekScore,
			"streak_days": sc.StreakDays,
		})
	}
	return result
}

// GetScoreLogs returns score change history for a user.
func (s *ScoreService) GetScoreLogs(userID uint, offset, limit int) []model.ScoreLog {
	var logs []model.ScoreLog
	s.db.Where("user_id = ?", userID).Order("created_at DESC").Offset(offset).Limit(limit).Find(&logs)
	return logs
}

// ShareCount returns the total number of shares by a user.
func (s *ScoreService) ShareCount(userID uint) int64 {
	var count int64
	s.db.Model(&model.ShareRecord{}).Where("user_id = ?", userID).Count(&count)
	return count
}

// InviteCount returns the number of users invited by a user.
func (s *ScoreService) InviteCount(userID uint) int64 {
	var count int64
	s.db.Model(&model.InviteRelation{}).Where("inviter_user_id = ?", userID).Count(&count)
	return count
}

// DishCount returns the number of user-created dishes.
func (s *ScoreService) DishCount(userID uint) int64 {
	var count int64
	s.db.Model(&model.Dish{}).Where("user_id = ? AND is_system = false", userID).Count(&count)
	return count
}

// SpinCount returns the total number of wheel spins by a user.
func (s *ScoreService) SpinCount(userID uint) int64 {
	var count int64
	s.db.Model(&model.ScoreLog{}).Where("user_id = ? AND action = ?", userID, "spin").Count(&count)
	return count
}

func (s *ScoreService) getOrCreateScore(userID uint) model.UserScore {
	var score model.UserScore
	if err := s.db.Where("user_id = ?", userID).First(&score).Error; err != nil {
		score = model.UserScore{UserID: userID}
		s.db.Create(&score)
	}
	return score
}
