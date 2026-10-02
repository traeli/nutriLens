package pgsql

import "testing"

func TestPersistentTableNamesRemainStable(t *testing.T) {
	tests := map[string]string{
		"user":                  (User{}).TableName(),
		"privacy agreement":     (PrivacyAgreement{}).TableName(),
		"city":                  (City{}).TableName(),
		"city poster theme":     (CityPosterTheme{}).TableName(),
		"city route":            (CityRoute{}).TableName(),
		"place":                 (Place{}).TableName(),
		"tag":                   (Tag{}).TableName(),
		"visit record":          (VisitRecord{}).TableName(),
		"visit record version":  (VisitRecordVersion{}).TableName(),
		"visit record tag link": (VisitRecordTagLink{}).TableName(),
		"record media":          (RecordMedia{}).TableName(),
		"record evidence":       (RecordEvidence{}).TableName(),
		"user trust profile":    (UserTrustProfile{}).TableName(),
		"contribution account":  (ContributionAccount{}).TableName(),
		"badge":                 (Badge{}).TableName(),
		"user badge":            (UserBadge{}).TableName(),
	}
	want := map[string]string{
		"user": "nutrilens_users", "privacy agreement": "privacy_agreements", "city": "cities", "city poster theme": "city_poster_themes", "city route": "city_routes", "place": "places", "tag": "tags",
		"visit record": "visit_records", "visit record version": "visit_record_versions", "visit record tag link": "visit_record_tag_links",
		"record media": "record_media", "record evidence": "record_evidences", "user trust profile": "user_trust_profiles",
		"contribution account": "contribution_accounts", "badge": "badges", "user badge": "user_badges",
	}
	for name, got := range tests {
		if got != want[name] {
			t.Errorf("%s table = %q, want %q", name, got, want[name])
		}
	}
}
