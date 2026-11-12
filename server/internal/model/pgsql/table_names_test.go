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
		"place theme":          (PlaceTheme{}).TableName(),
		"tag":                  (Tag{}).TableName(),
		"visit record":         (VisitRecord{}).TableName(),
		"restaurant review":    (RestaurantReview{}).TableName(),
		"review version":       (RestaurantReviewVersion{}).TableName(),
		"review tag link":      (RestaurantReviewTagLink{}).TableName(),
		"discovery post":       (DiscoveryPost{}).TableName(),
		"discovery post image": (DiscoveryPostImage{}).TableName(),
		"discovery post like":  (DiscoveryPostLike{}).TableName(),
		"nutrition record":     (NutritionRecord{}).TableName(),
		"user trust profile":   (UserTrustProfile{}).TableName(),
		"contribution account": (ContributionAccount{}).TableName(),
		"badge":                (Badge{}).TableName(),
		"user badge":           (UserBadge{}).TableName(),
	}
	want := map[string]string{
		"user": "nutrilens_users", "privacy agreement": "privacy_agreements", "city": "cities", "city poster theme": "city_poster_themes", "city route": "city_routes", "place": "places", "place theme": "place_themes", "tag": "tags",
		"visit record": "visit_records", "user trust profile": "user_trust_profiles",
		"restaurant review": "restaurant_reviews", "review version": "restaurant_review_versions",
		"review tag link": "restaurant_review_tag_links", "nutrition record": "nutrition_records",
		"discovery post": "discovery_posts", "discovery post image": "discovery_post_images", "discovery post like": "discovery_post_likes",
		"contribution account": "contribution_accounts", "badge": "badges", "user badge": "user_badges",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s table = %q, want %q", name, got, want[name])
		}
	}
}
