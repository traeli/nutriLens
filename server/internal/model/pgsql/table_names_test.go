package pgsql

import "testing"

func TestPersistentTableNamesRemainStable(t *testing.T) {
	tests := map[string]string{
		"user":                 (User{}).TableName(),
		"privacy agreement":    (PrivacyAgreement{}).TableName(),
		"city":                 (City{}).TableName(),
		"city poster theme":    (CityPosterTheme{}).TableName(),
		"city route":           (CityRoute{}).TableName(),
		"place":                (Place{}).TableName(),
		"tag":                  (Tag{}).TableName(),
		"visit record":         (VisitRecord{}).TableName(),
		"restaurant review":    (RestaurantReview{}).TableName(),
		"review version":       (RestaurantReviewVersion{}).TableName(),
		"review tag link":      (RestaurantReviewTagLink{}).TableName(),
		"nutrition record":     (NutritionRecord{}).TableName(),
		"user trust profile":   (UserTrustProfile{}).TableName(),
		"contribution account": (ContributionAccount{}).TableName(),
		"badge":                (Badge{}).TableName(),
		"user badge":           (UserBadge{}).TableName(),
	}
	want := map[string]string{
		"user": "nutrilens_users", "privacy agreement": "privacy_agreements", "city": "cities", "city poster theme": "city_poster_themes", "city route": "city_routes", "place": "places", "tag": "tags",
		"visit record": "visit_records", "user trust profile": "user_trust_profiles",
		"restaurant review": "restaurant_reviews", "review version": "restaurant_review_versions",
		"review tag link": "restaurant_review_tag_links", "nutrition record": "nutrition_records",
		"contribution account": "contribution_accounts", "badge": "badges", "user badge": "user_badges",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s table = %q, want %q", name, got, want[name])
		}
	}
}
