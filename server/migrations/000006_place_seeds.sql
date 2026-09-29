-- Seed catalog places for the search page ("找一家店").
-- Idempotent: each row is skipped when a place with the same name already
-- exists in the city (user-submitted places take precedence over seeds).
BEGIN;

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '老周生煎', 'stall', '310000', '静安区', '曹家渡', '长宁路123号', 121.4288000, 31.2282000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '310000' AND name = '老周生煎');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '阿庆牛肉面', 'restaurant', '310000', '静安区', '曹家渡', '万航渡路245号', 121.4305000, 31.2275000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '310000' AND name = '阿庆牛肉面');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '弄堂豆浆', 'stall', '310000', '静安区', '曹家渡', '万航渡路170号', 121.4298000, 31.2266000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '310000' AND name = '弄堂豆浆');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT 'Mono Coffee', 'drink', '310000', '静安区', '曹家渡', '长宁路88号', 121.4276000, 31.2290000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '310000' AND name = 'Mono Coffee');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '万寿斋', 'restaurant', '310000', '虹口区', '山阴路', '山阴路123号', 121.4749000, 31.2647000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '310000' AND name = '万寿斋');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '阿娘黄鱼面', 'restaurant', '310000', '黄浦区', '进贤路', '进贤路120号', 121.4678000, 31.2228000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '310000' AND name = '阿娘黄鱼面');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '菊英面店', 'restaurant', '330100', '上城区', '中山北路', '中山北路436号', 120.1722000, 30.2543000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '330100' AND name = '菊英面店');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '方老大面馆', 'restaurant', '330100', '上城区', '鼓楼', '中山南路77号', 120.1738000, 30.2437000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '330100' AND name = '方老大面馆');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '游埠豆浆', 'stall', '330100', '上城区', '城头巷', '城头巷101号', 120.1745000, 30.2478000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '330100' AND name = '游埠豆浆');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '知味观', 'restaurant', '330100', '上城区', '湖滨', '仁和路83号', 120.1661000, 30.2494000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '330100' AND name = '知味观');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '老贯桥烧麦', 'stall', '330100', '拱墅区', '武林门', '老贯桥巷8号', 120.1623000, 30.2791000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '330100' AND name = '老贯桥烧麦');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '老头儿油爆虾', 'restaurant', '330100', '上城区', '湖滨', '邮电路46号', 120.1643000, 30.2486000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '330100' AND name = '老头儿油爆虾');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '同得兴', 'restaurant', '320500', '姑苏区', '观前街', '嘉馀坊6号', 120.6178000, 31.3165000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320500' AND name = '同得兴');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '裕兴记', 'restaurant', '320500', '姑苏区', '观前街', '白塔西路60号', 120.6233000, 31.3224000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320500' AND name = '裕兴记');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '哑巴生煎', 'stall', '320500', '姑苏区', '平江路', '临顿路温家岸12号', 120.6300000, 31.3198000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320500' AND name = '哑巴生煎');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '朱鸿兴', 'restaurant', '320500', '姑苏区', '观前街', '碧凤坊13号', 120.6193000, 31.3152000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320500' AND name = '朱鸿兴');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '琼琳阁面庄', 'restaurant', '320500', '姑苏区', '十全街', '书院巷20号', 120.6228000, 31.2952000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320500' AND name = '琼琳阁面庄');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '双塔市集', 'night_market', '320500', '姑苏区', '双塔', '石匠弄2号', 120.6268000, 31.2988000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320500' AND name = '双塔市集');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '李记清真馆', 'stall', '320100', '秦淮区', '评事街', '打钉巷1号', 118.7735000, 32.0129000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320100' AND name = '李记清真馆');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '鸡鸣汤包', 'restaurant', '320100', '鼓楼区', '湖南路', '狮子桥美食街', 118.7787000, 32.0663000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320100' AND name = '鸡鸣汤包');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '芳婆糕团店', 'stall', '320100', '秦淮区', '王府大街', '王府大街138号', 118.7743000, 32.0398000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320100' AND name = '芳婆糕团店');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '蒋有记', 'stall', '320100', '秦淮区', '老门东', '箍桶巷三条营', 118.7892000, 32.0096000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320100' AND name = '蒋有记');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '章云板鸭', 'restaurant', '320100', '雨花台区', '雨花西路', '雨花西路2号', 118.7682000, 31.9905000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320100' AND name = '章云板鸭');

INSERT INTO places (name, category, city_code, district, business_area, address, longitude, latitude, status)
SELECT '蓝老大糖粥藕', 'stall', '320100', '秦淮区', '夫子庙', '双塘园18号', 118.7808000, 32.0065000, 'active'
WHERE NOT EXISTS (SELECT 1 FROM places WHERE city_code = '320100' AND name = '蓝老大糖粥藕');

COMMIT;
