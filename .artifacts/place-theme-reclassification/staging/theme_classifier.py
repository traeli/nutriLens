from __future__ import annotations

import re


VALID_THEME_CODES = frozenset({
    "dumpling_bun", "noodles", "rice_meal", "local_dishes", "hotpot",
    "bbq_skewers", "street_snack", "breakfast", "coffee_tea", "dessert",
    "bakery", "seafood", "vegetarian", "bar_drinks", "general",
})

_NAME_RULES = (
    ("coffee_tea", ("咖啡", "coffee", "茶颜", "茶饮", "奶茶")),
    ("dessert", ("甜品", "糖水", "糖粥", "甜食", "双皮奶", "清补凉", "醪糟", "冰品", "炒冰")),
    ("bakery", ("面包", "烘焙", "西饼", "蛋糕", "月饼", "列巴")),
    ("vegetarian", ("素食", "轻食", "沙拉", "斋菜")),
    ("hotpot", ("火锅", "涮肉", "串串", "鸡公煲", "猪骨煲", "砂锅", "铁锅炖", "菌锅", "美蛙鱼头")),
    ("bbq_skewers", ("烧烤", "烤肉", "烧肉", "烧鸟", "烤串", "烤兔", "烧骨", "烤鱼")),
    ("bar_drinks", ("酒吧", "酒馆", "精酿", "居酒屋", "酒食")),
    ("dumpling_bun", ("汤包", "小笼", "生煎", "烧麦", "锅贴", "煎饺", "水饺", "饺子", "虾饺", "点心", "茶点", "茶楼", "包子")),
    ("noodles", ("面馆", "面店", "面庄", "牛肉面", "黄鱼面", "蟹黄面", "裤带面", "biangbiang", "拌面", "馄饨", "粉丝汤", "米粉", "粉店", "面粉馆", "米线", "凉粉", "豆丝", "泡馍")),
    ("rice_meal", ("煲仔饭", "盖饭", "便当", "饭团", "手抓饭")),
    ("breakfast", ("早餐", "早点", "豆浆", "油条", "粥铺", "粥店", "番薯粥")),
    ("seafood", ("海鲜", "河鲜", "渔船", "煎蟹", "小龙虾", "虾庄", "虾一跳", "油爆虾", "寿司")),
    ("street_snack", ("小吃", "夜市", "市集", "好吃街", "臭豆腐", "肉夹馍", "凉皮", "蛋烘糕", "糕团", "烧肉粽")),
)

_NON_PLACE_MARKERS = ("人民政府", "管理委员会", "办事处", "停车场")


def _first_rule(text: str) -> str | None:
    for code, keywords in _NAME_RULES:
        if any(keyword in text for keyword in keywords):
            return code
    if re.search(r"(?:面|粉)(?:[·（(]|$)", text):
        return "noodles"
    return None


def infer_theme_code(
    *,
    name: str,
    category: str = "",
    description: str = "",
    current: str = "",
) -> str:
    """Infer a stable route theme while preserving an explicit non-general value."""
    current = str(current or "").strip()
    if current in VALID_THEME_CODES and current != "general":
        return current

    name_text = str(name or "").strip().lower()
    category = str(category or "").strip().lower()
    if any(marker in name_text for marker in _NON_PLACE_MARKERS) and category != "night_market":
        return "general"

    inferred = _first_rule(name_text)
    if inferred:
        return inferred
    if category == "night_market":
        return "street_snack"
    if category == "stall":
        return "street_snack"

    inferred = _first_rule(str(description or "").strip().lower())
    if inferred:
        return inferred
    if category == "drink":
        return "coffee_tea"
    if category == "restaurant":
        return "local_dishes"
    return "general"
