import unittest

from food_route_collector.theme_classifier import infer_theme_code


class ThemeClassifierTests(unittest.TestCase):
    def test_infers_representative_themes(self):
        cases = (
            ("阿娘黄鱼面", "restaurant", "noodles"),
            ("鸡鸣汤包", "restaurant", "dumpling_bun"),
            ("巴奴毛肚火锅", "restaurant", "hotpot"),
            ("盟重烧烤", "restaurant", "bbq_skewers"),
            ("癸卯GUIMOU COFFEE", "drink", "coffee_tea"),
            ("第一市场夜市", "night_market", "street_snack"),
            ("遇见便当", "restaurant", "rice_meal"),
        )
        for name, category, expected in cases:
            with self.subTest(name=name):
                self.assertEqual(infer_theme_code(name=name, category=category), expected)

    def test_preserves_explicit_theme_and_handles_fallbacks(self):
        self.assertEqual(infer_theme_code(name="任意店", category="restaurant", current="seafood"), "seafood")
        self.assertEqual(infer_theme_code(name="苏州市吴中区人民政府", category="restaurant"), "general")
        self.assertEqual(infer_theme_code(name="社区家常馆", category="restaurant"), "local_dishes")


if __name__ == "__main__":
    unittest.main()
