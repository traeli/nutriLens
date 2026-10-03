package pgsql

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Seed 写入新环境运行所需的基础产品数据；遇到冲突时保留已有数据，避免覆盖运营配置。
func Seed(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := seedCityCatalog(tx); err != nil {
			return err
		}
		if err := seedPlaces(tx); err != nil {
			return err
		}
		if err := createSeed(tx, "poster themes", seedPosterThemes()); err != nil {
			return err
		}
		if err := createSeed(tx, "experience tags", seedTags()); err != nil {
			return err
		}
		if err := createSeed(tx, "city routes", seedRoutes()); err != nil {
			return err
		}
		if err := seedRouteStops(tx); err != nil {
			return err
		}
		if err := createSeed(tx, "feature flags", seedFeatureFlags()); err != nil {
			return err
		}
		return nil
	})
}

func seedCityCatalog(tx *gorm.DB) error {
	var defaultCount int64
	if err := tx.Model(&City{}).Where("is_default = ?", true).Count(&defaultCount).Error; err != nil {
		return fmt.Errorf("inspect default city: %w", err)
	}
	cities := seedCities()
	if defaultCount > 0 {
		cities[0].IsDefault = false
	}
	for _, city := range cities {
		if err := createSeed(tx, "city "+city.Code, &city); err != nil {
			return err
		}
	}
	return nil
}

func createSeed(tx *gorm.DB, name string, values any) error {
	if err := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(values).Error; err != nil {
		return fmt.Errorf("seed %s: %w", name, err)
	}
	return nil
}

func seedPlaces(tx *gorm.DB) error {
	places := []Place{
		{Name: "老周生煎", Category: "stall", CityCode: "310000", District: "静安区", BusinessArea: "曹家渡", Address: "长宁路123号", Longitude: 121.4288, Latitude: 31.2282, Status: "active"},
		{Name: "阿庆牛肉面", Category: "restaurant", CityCode: "310000", District: "静安区", BusinessArea: "曹家渡", Address: "万航渡路245号", Longitude: 121.4305, Latitude: 31.2275, Status: "active"},
		{Name: "弄堂豆浆", Category: "stall", CityCode: "310000", District: "静安区", BusinessArea: "曹家渡", Address: "万航渡路170号", Longitude: 121.4298, Latitude: 31.2266, Status: "active"},
		{Name: "Mono Coffee", Category: "drink", CityCode: "310000", District: "静安区", BusinessArea: "曹家渡", Address: "长宁路88号", Longitude: 121.4276, Latitude: 31.229, Status: "active"},
		{Name: "万寿斋", Category: "restaurant", CityCode: "310000", District: "虹口区", BusinessArea: "山阴路", Address: "山阴路123号", Longitude: 121.4749, Latitude: 31.2647, Status: "active"},
		{Name: "阿娘黄鱼面", Category: "restaurant", CityCode: "310000", District: "黄浦区", BusinessArea: "进贤路", Address: "进贤路120号", Longitude: 121.4678, Latitude: 31.2228, Status: "active"},
		{Name: "菊英面店", Category: "restaurant", CityCode: "330100", District: "上城区", BusinessArea: "中山北路", Address: "中山北路436号", Longitude: 120.1722, Latitude: 30.2543, Status: "active"},
		{Name: "方老大面馆", Category: "restaurant", CityCode: "330100", District: "上城区", BusinessArea: "鼓楼", Address: "中山南路77号", Longitude: 120.1738, Latitude: 30.2437, Status: "active"},
		{Name: "游埠豆浆", Category: "stall", CityCode: "330100", District: "上城区", BusinessArea: "城头巷", Address: "城头巷101号", Longitude: 120.1745, Latitude: 30.2478, Status: "active"},
		{Name: "知味观", Category: "restaurant", CityCode: "330100", District: "上城区", BusinessArea: "湖滨", Address: "仁和路83号", Longitude: 120.1661, Latitude: 30.2494, Status: "active"},
		{Name: "老贯桥烧麦", Category: "stall", CityCode: "330100", District: "拱墅区", BusinessArea: "武林门", Address: "老贯桥巷8号", Longitude: 120.1623, Latitude: 30.2791, Status: "active"},
		{Name: "老头儿油爆虾", Category: "restaurant", CityCode: "330100", District: "上城区", BusinessArea: "湖滨", Address: "邮电路46号", Longitude: 120.1643, Latitude: 30.2486, Status: "active"},
		{Name: "同得兴", Category: "restaurant", CityCode: "320500", District: "姑苏区", BusinessArea: "观前街", Address: "嘉馀坊6号", Longitude: 120.6178, Latitude: 31.3165, Status: "active"},
		{Name: "裕兴记", Category: "restaurant", CityCode: "320500", District: "姑苏区", BusinessArea: "观前街", Address: "白塔西路60号", Longitude: 120.6233, Latitude: 31.3224, Status: "active"},
		{Name: "哑巴生煎", Category: "stall", CityCode: "320500", District: "姑苏区", BusinessArea: "平江路", Address: "临顿路温家岸12号", Longitude: 120.63, Latitude: 31.3198, Status: "active"},
		{Name: "朱鸿兴", Category: "restaurant", CityCode: "320500", District: "姑苏区", BusinessArea: "观前街", Address: "碧凤坊13号", Longitude: 120.6193, Latitude: 31.3152, Status: "active"},
		{Name: "琼琳阁面庄", Category: "restaurant", CityCode: "320500", District: "姑苏区", BusinessArea: "十全街", Address: "书院巷20号", Longitude: 120.6228, Latitude: 31.2952, Status: "active"},
		{Name: "双塔市集", Category: "night_market", CityCode: "320500", District: "姑苏区", BusinessArea: "双塔", Address: "石匠弄2号", Longitude: 120.6268, Latitude: 31.2988, Status: "active"},
		{Name: "李记清真馆", Category: "stall", CityCode: "320100", District: "秦淮区", BusinessArea: "评事街", Address: "打钉巷1号", Longitude: 118.7735, Latitude: 32.0129, Status: "active"},
		{Name: "鸡鸣汤包", Category: "restaurant", CityCode: "320100", District: "鼓楼区", BusinessArea: "湖南路", Address: "狮子桥美食街", Longitude: 118.7787, Latitude: 32.0663, Status: "active"},
		{Name: "芳婆糕团店", Category: "stall", CityCode: "320100", District: "秦淮区", BusinessArea: "王府大街", Address: "王府大街138号", Longitude: 118.7743, Latitude: 32.0398, Status: "active"},
		{Name: "蒋有记", Category: "stall", CityCode: "320100", District: "秦淮区", BusinessArea: "老门东", Address: "箍桶巷三条营", Longitude: 118.7892, Latitude: 32.0096, Status: "active"},
		{Name: "章云板鸭", Category: "restaurant", CityCode: "320100", District: "雨花台区", BusinessArea: "雨花西路", Address: "雨花西路2号", Longitude: 118.7682, Latitude: 31.9905, Status: "active"},
		{Name: "蓝老大糖粥藕", Category: "stall", CityCode: "320100", District: "秦淮区", BusinessArea: "夫子庙", Address: "双塘园18号", Longitude: 118.7808, Latitude: 32.0065, Status: "active"},
	}
	for _, place := range places {
		if err := tx.Where("city_code = ? AND name = ?", place.CityCode, place.Name).FirstOrCreate(&place).Error; err != nil {
			return fmt.Errorf("seed place %s/%s: %w", place.CityCode, place.Name, err)
		}
	}
	return nil
}

func seedRouteStops(tx *gorm.DB) error {
	var route CityRoute
	if err := tx.Where("city_code = ? AND title = ?", "310000", "曹家渡的四站早餐").First(&route).Error; err != nil {
		return fmt.Errorf("load seeded route: %w", err)
	}
	for index, name := range []string{"老周生煎", "阿庆牛肉面", "弄堂豆浆", "Mono Coffee"} {
		var place Place
		if err := tx.Where("city_code = ? AND name = ?", route.CityCode, name).First(&place).Error; err != nil {
			return fmt.Errorf("load seeded route place %s: %w", name, err)
		}
		stop := CityRouteStop{RouteID: route.ID, PlaceID: place.ID, SortOrder: index + 1}
		if err := tx.Where("route_id = ? AND sort_order = ?", route.ID, index+1).FirstOrCreate(&stop).Error; err != nil {
			return fmt.Errorf("seed route stop %s: %w", name, err)
		}
	}
	return nil
}

func seedCities() []City {
	return []City{
		{Code: "310000", Name: "上海", Description: "在巷子里 遇见生活", Image: "/static/dining/shanghai-city.jpg", Enabled: true, IsDefault: true, SortOrder: 10},
		{Code: "330100", Name: "杭州", Description: "沿湖而行 吃进四季", Image: "/static/dining/corner-cafe.jpg", Enabled: true, SortOrder: 20},
		{Code: "320500", Name: "苏州", Description: "转进小巷 尝一口江南", Image: "/static/dining/noodle-shop.jpg", Enabled: true, SortOrder: 30},
		{Code: "320100", Name: "南京", Description: "顺着城墙 找老味道", Image: "/static/dining/noodle-detail.jpg", Enabled: true, SortOrder: 40},
	}
}

func seedPosterThemes() []CityPosterTheme {
	return []CityPosterTheme{
		{CityCode: "310000", Version: 1, Title: "上海食光星图", Subtitle: "梧桐影里，记下认真吃饭的夜晚", BackgroundImage: "/static/posters/food-memory-night-v1.jpg", AccentColor: "#C7FF35", SecondaryColor: "#F1E5C8", Motifs: []byte(`["梧桐","弄堂","夜色"]`)},
		{CityCode: "330100", Version: 1, Title: "杭州食光星图", Subtitle: "风经过湖面，也经过一餐一饭", BackgroundImage: "/static/posters/food-memory-night-v1.jpg", AccentColor: "#BDEB7D", SecondaryColor: "#F2E7C9", Motifs: []byte(`["桂花","晚风","茶香"]`)},
		{CityCode: "320500", Version: 1, Title: "苏州食光星图", Subtitle: "把巷口与热汤，收进生活的星河", BackgroundImage: "/static/posters/food-memory-night-v1.jpg", AccentColor: "#D5F56A", SecondaryColor: "#E9DDC3", Motifs: []byte(`["小巷","水声","热汤"]`)},
		{CityCode: "320100", Version: 1, Title: "南京食光星图", Subtitle: "晚风、灯火，以及记得住的味道", BackgroundImage: "/static/posters/food-memory-night-v1.jpg", AccentColor: "#D7FF4A", SecondaryColor: "#EED9BE", Motifs: []byte(`["城墙","晚风","灯火"]`)},
	}
}

func seedTags() []Tag {
	return []Tag{
		{Code: "taste", Name: "口味", GroupName: "experience", Enabled: true, SortOrder: 10},
		{Code: "price", Name: "价格", GroupName: "experience", Enabled: true, SortOrder: 20},
		{Code: "service", Name: "服务", GroupName: "experience", Enabled: true, SortOrder: 30},
		{Code: "queue", Name: "排队", GroupName: "experience", Enabled: true, SortOrder: 40},
		{Code: "hygiene_observation", Name: "卫生观感", GroupName: "experience", Enabled: true, SortOrder: 50},
		{Code: "promotion_mismatch", Name: "宣传不符", GroupName: "experience", Enabled: true, SortOrder: 60},
		{Code: "other", Name: "其他", GroupName: "experience", Enabled: true, SortOrder: 99},
	}
}

func seedRoutes() []CityRoute {
	return []CityRoute{
		{CityCode: "310000", Title: "曹家渡的四站早餐", Tag: "早餐路线", Meta: "2.5小时 · 步行2.8km", Keyword: "曹家渡", Image: "/static/dining/noodle-shop.jpg", Footnote: "从热气腾腾，到一杯咖啡", Stops: JSONDocument(`["老周生煎","阿庆牛肉面","弄堂豆浆","Mono Coffee"]`), IsFeatured: true, Enabled: true, SortOrder: 10},
		{CityCode: "310000", Title: "下班后还亮着灯的社区小馆", Tag: "小巷路线", Meta: "静安 · 晚餐 · 6 家", Keyword: "南京西路", Image: "/static/dining/corner-cafe.jpg", Stops: JSONDocument(`[]`), Enabled: true, SortOrder: 20},
		{CityCode: "330100", Title: "从西湖边拐进大井巷", Tag: "小巷路线", Meta: "1.8小时 · 步行1.6km", Keyword: "大井巷", Image: "/static/dining/corner-cafe.jpg", Stops: JSONDocument(`[]`), Enabled: true, SortOrder: 10},
		{CityCode: "320500", Title: "从平江路拐进寻常巷陌", Tag: "小巷路线", Meta: "1.8小时 · 步行1.6km", Keyword: "平江路", Image: "/static/dining/noodle-detail.jpg", Stops: JSONDocument(`[]`), Enabled: true, SortOrder: 10},
		{CityCode: "320100", Title: "从一碗面开始认识老门东", Tag: "早餐路线", Meta: "2小时 · 步行2.2km", Keyword: "老门东", Image: "/static/dining/noodle-shop.jpg", Stops: JSONDocument(`[]`), Enabled: true, SortOrder: 10},
	}
}

func seedFeatureFlags() []FeatureFlag {
	return []FeatureFlag{
		{Key: "public_submission_enabled", Enabled: true, Value: JSONDocument(`{}`), Description: "公开评论提交总开关"},
		{Key: "media_upload_enabled", Enabled: true, Value: JSONDocument(`{}`), Description: "评论图片和私密消费凭证上传开关"},
	}
}
