package service

import (
	"fmt"
	"log"
	"time"

	"nutrilens/internal/model"

	"gorm.io/gorm"
)

type AchievementService struct {
	db    *gorm.DB
	score *ScoreService
}

func NewAchievementService(db *gorm.DB, scoreSvc *ScoreService) *AchievementService {
	return &AchievementService{db: db, score: scoreSvc}
}

// CheckAndUnlock checks all achievements for a user and unlocks any newly earned ones.
// Returns the list of newly unlocked achievements.
func (a *AchievementService) CheckAndUnlock(userID uint) []model.Achievement {
	var achievements []model.Achievement
	a.db.Where("enabled = true").Find(&achievements)

	// Get already unlocked
	var unlocked []model.UserAchievement
	a.db.Where("user_id = ?", userID).Find(&unlocked)
	unlockedMap := map[uint]bool{}
	for _, u := range unlocked {
		unlockedMap[u.AchievementID] = true
	}

	// Count stats
	totalAnalyzes := int(a.totalAnalyzes(userID))
	streakDays := a.currentStreak(userID)
	shareCount := int(a.score.ShareCount(userID))
	inviteCount := int(a.score.InviteCount(userID))
	dishCount := int(a.score.DishCount(userID))
	spinCount := int(a.score.SpinCount(userID))
	threeMealCount := int(a.threeMealCount(userID))
	healthyDays := a.healthyStreak(userID)
	commonRareCount := a.commonRareUnlockedCount(userID)

	var newUnlocks []model.Achievement

	for _, ach := range achievements {
		if unlockedMap[ach.ID] {
			continue
		}

		unlocked := false
		switch ach.ConditionType {
		case "first_analyze":
			unlocked = totalAnalyzes >= 1
		case "total_analyze":
			unlocked = totalAnalyzes >= ach.ConditionValue
		case "streak":
			unlocked = streakDays >= ach.ConditionValue
		case "healthy_streak":
			unlocked = healthyDays >= ach.ConditionValue
		case "three_meals":
			unlocked = threeMealCount >= ach.ConditionValue
		case "spin":
			unlocked = spinCount >= ach.ConditionValue
		case "dish":
			unlocked = dishCount >= ach.ConditionValue
		case "share":
			unlocked = shareCount >= ach.ConditionValue
		case "invite":
			unlocked = inviteCount >= ach.ConditionValue
		case "all_common_rare":
			unlocked = commonRareCount >= ach.ConditionValue
		}

		if unlocked {
			ua := model.UserAchievement{
				UserID:        userID,
				AchievementID: ach.ID,
				UnlockedAt:    time.Now(),
			}
			if err := a.db.Create(&ua).Error; err != nil {
				log.Printf("[Achievement] failed to unlock: user=%d, achievement=%s, err=%v", userID, ach.Code, err)
				continue
			}
			newUnlocks = append(newUnlocks, ach)

			// Apply rewards
			a.applyReward(userID, ach)
		}
	}

	return newUnlocks
}

// ListAchievements returns all achievements with unlock status for a user.
func (a *AchievementService) ListAchievements(userID uint) []map[string]interface{} {
	var achievements []model.Achievement
	a.db.Where("enabled = true").Order("id ASC").Find(&achievements)

	var unlocked []model.UserAchievement
	a.db.Where("user_id = ?", userID).Find(&unlocked)
	unlockedMap := map[uint]time.Time{}
	for _, u := range unlocked {
		unlockedMap[u.AchievementID] = u.UnlockedAt
	}

	// Get user's equipped title
	var user model.User
	a.db.Select("title").First(&user, userID)
	equippedTitleID := user.Title

	var result []map[string]interface{}
	for _, ach := range achievements {
		unlockedAt, isUnlocked := unlockedMap[ach.ID]
		item := map[string]interface{}{
			"id":              ach.ID,
			"code":            ach.Code,
			"name":            ach.Name,
			"icon":            ach.Icon,
			"description":     ach.Description,
			"rarity":          ach.Rarity,
			"condition_type":  ach.ConditionType,
			"condition_value": ach.ConditionValue,
			"reward_type":     ach.RewardType,
			"reward_value":    ach.RewardValue,
			"unlocked":        isUnlocked,
			"unlocked_at":     unlockedAt,
			"is_equipped":     isUnlocked && ach.ID == equippedTitleID,
		}
		result = append(result, item)
	}
	return result
}

// SetTitle sets the user's currently displayed achievement title.
func (a *AchievementService) SetTitle(userID, achievementID uint) error {
	// Verify user has unlocked this achievement
	var count int64
	a.db.Model(&model.UserAchievement{}).
		Where("user_id = ? AND achievement_id = ?", userID, achievementID).
		Count(&count)
	if count == 0 {
		return fmt.Errorf("achievement not unlocked")
	}

	return a.db.Model(&model.User{}).Where("id = ?", userID).
		Update("title", achievementID).Error
}

func (a *AchievementService) applyReward(userID uint, ach model.Achievement) {
	switch ach.RewardType {
	case "quota":
		// Add extra AI quota to user score
		var score model.UserScore
		if err := a.db.Where("user_id = ?", userID).First(&score).Error; err != nil {
			return
		}
		a.db.Model(&score).Update("bonus_quota", score.BonusQuota+ach.RewardValue)
		log.Printf("[Achievement] applied quota reward: user=%d, +=%d", userID, ach.RewardValue)

	case "vip":
		a.db.Model(&model.User{}).Where("id = ?", userID).Update("tag", "vip")
		log.Printf("[Achievement] applied VIP reward: user=%d", userID)

	case "poster_template":
		log.Printf("[Achievement] applied poster template reward: user=%d, template=%d", userID, ach.RewardValue)
	}
}

func (a *AchievementService) totalAnalyzes(userID uint) int64 {
	var count int64
	a.db.Model(&model.AICallLog{}).
		Where("user_id = ? AND call_type IN ?", userID, []string{"analyze_text", "analyze_image"}).
		Count(&count)
	return count
}

func (a *AchievementService) currentStreak(userID uint) int {
	var score model.UserScore
	if err := a.db.Where("user_id = ?", userID).First(&score).Error; err != nil {
		return 0
	}
	return score.StreakDays
}

func (a *AchievementService) threeMealCount(userID uint) int64 {
	// Count distinct dates where user has all 3 meal types
	var count int64
	a.db.Raw(`
		SELECT COUNT(DISTINCT d::date) FROM (
			SELECT created_at::date AS d, COUNT(DISTINCT meal_type) AS mc
			FROM food_records
			WHERE user_id = ? AND meal_type IN (1, 2, 3) AND deleted_at IS NULL
			GROUP BY created_at::date
		) sub WHERE mc >= 3
	`, userID).Scan(&count)
	return count
}

func (a *AchievementService) healthyStreak(userID uint) int {
	// Count consecutive days (ending today or yesterday) where total calories is in a healthy range
	// Healthy range: roughly 1200-2200 kcal (depends on gender, but simplified)
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	streak := 0
	for i := 0; i < 365; i++ {
		day := today.AddDate(0, 0, -i)
		nextDay := day.AddDate(0, 0, 1)

		var totalCal float64
		a.db.Model(&model.FoodRecord{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, day, nextDay).
			Select("COALESCE(SUM(calories), 0)").
			Scan(&totalCal)

		if totalCal >= 1200 && totalCal <= 2200 {
			streak++
		} else if i == 0 {
			// Today might not be over yet, skip
			continue
		} else {
			break
		}
	}
	return streak
}

func (a *AchievementService) commonRareUnlockedCount(userID uint) int {
	var count int64
	a.db.Model(&model.UserAchievement{}).
		Joins("JOIN achievements ON achievements.id = user_achievements.achievement_id").
		Where("user_achievements.user_id = ? AND achievements.rarity IN ?", userID, []string{"common", "rare"}).
		Count(&count)
	return int(count)
}
