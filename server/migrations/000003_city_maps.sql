BEGIN;

CREATE TABLE IF NOT EXISTS city_maps (
    city_code VARCHAR(16) PRIMARY KEY REFERENCES cities(code) ON DELETE CASCADE,
    version INTEGER NOT NULL DEFAULT 1 CHECK (version > 0),
    center_longitude DOUBLE PRECISION NOT NULL,
    center_latitude DOUBLE PRECISION NOT NULL,
    min_longitude DOUBLE PRECISION NOT NULL,
    min_latitude DOUBLE PRECISION NOT NULL,
    max_longitude DOUBLE PRECISION NOT NULL,
    max_latitude DOUBLE PRECISION NOT NULL,
    geometry JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CHECK (min_longitude < max_longitude),
    CHECK (min_latitude < max_latitude),
    CHECK (jsonb_typeof(geometry) = 'object')
);

COMMENT ON TABLE city_maps IS '用于小程序城市星图画布的轻量归一化几何数据';
COMMENT ON COLUMN city_maps.geometry IS '归一化到 0 至 1 范围的轮廓、水系和简化道路路径';

INSERT INTO city_maps (
    city_code, version, center_longitude, center_latitude,
    min_longitude, min_latitude, max_longitude, max_latitude, geometry
) VALUES
(
    '310000', 1, 121.4737, 31.2304, 120.85, 30.67, 122.12, 31.88,
    $json${
      "outline":[[0.08,0.22],[0.18,0.10],[0.42,0.07],[0.62,0.13],[0.78,0.20],[0.91,0.38],[0.88,0.57],[0.78,0.73],[0.60,0.89],[0.39,0.91],[0.20,0.80],[0.09,0.61],[0.05,0.40],[0.08,0.22]],
      "waterways":[[[0.07,0.48],[0.21,0.45],[0.34,0.49],[0.47,0.45],[0.61,0.39],[0.74,0.33],[0.91,0.31]],[[0.62,0.13],[0.65,0.28],[0.61,0.39],[0.64,0.57],[0.72,0.75]]],
      "roads":[
        {"level":"primary","points":[[0.12,0.67],[0.29,0.58],[0.47,0.51],[0.68,0.46],[0.87,0.42]]},
        {"level":"primary","points":[[0.22,0.14],[0.31,0.30],[0.43,0.46],[0.55,0.63],[0.66,0.84]]},
        {"level":"primary","points":[[0.13,0.34],[0.31,0.37],[0.49,0.35],[0.67,0.27],[0.82,0.21]]},
        {"level":"primary","points":[[0.22,0.81],[0.36,0.67],[0.51,0.58],[0.69,0.57],[0.83,0.66]]},
        {"level":"secondary","points":[[0.18,0.23],[0.26,0.41],[0.28,0.59],[0.31,0.78]]},
        {"level":"secondary","points":[[0.41,0.10],[0.40,0.29],[0.43,0.46],[0.42,0.70],[0.47,0.88]]},
        {"level":"secondary","points":[[0.57,0.15],[0.54,0.32],[0.55,0.49],[0.57,0.70],[0.60,0.86]]},
        {"level":"secondary","points":[[0.74,0.22],[0.68,0.39],[0.69,0.57],[0.76,0.72]]},
        {"level":"secondary","points":[[0.11,0.55],[0.28,0.51],[0.46,0.53],[0.63,0.62],[0.81,0.77]]},
        {"level":"secondary","points":[[0.18,0.70],[0.35,0.73],[0.55,0.72],[0.75,0.64]]}
      ]
    }$json$::jsonb
),
(
    '330100', 1, 120.1551, 30.2741, 119.30, 29.70, 120.72, 30.63,
    $json${
      "outline":[[0.12,0.18],[0.31,0.08],[0.57,0.09],[0.79,0.19],[0.92,0.36],[0.86,0.57],[0.72,0.74],[0.51,0.91],[0.28,0.86],[0.12,0.70],[0.06,0.43],[0.12,0.18]],
      "waterways":[[[0.08,0.34],[0.26,0.37],[0.43,0.45],[0.60,0.48],[0.78,0.43],[0.91,0.35]],[[0.34,0.72],[0.43,0.62],[0.45,0.49],[0.41,0.36],[0.35,0.20]]],
      "roads":[
        {"level":"primary","points":[[0.13,0.65],[0.29,0.56],[0.46,0.51],[0.65,0.52],[0.84,0.61]]},
        {"level":"primary","points":[[0.18,0.23],[0.34,0.34],[0.49,0.48],[0.62,0.67],[0.72,0.82]]},
        {"level":"primary","points":[[0.13,0.41],[0.30,0.43],[0.48,0.40],[0.68,0.32],[0.84,0.25]]},
        {"level":"primary","points":[[0.23,0.79],[0.38,0.66],[0.55,0.59],[0.74,0.57]]},
        {"level":"secondary","points":[[0.26,0.12],[0.28,0.31],[0.30,0.51],[0.28,0.74]]},
        {"level":"secondary","points":[[0.48,0.10],[0.47,0.28],[0.49,0.48],[0.48,0.69],[0.51,0.88]]},
        {"level":"secondary","points":[[0.68,0.14],[0.62,0.31],[0.62,0.50],[0.66,0.72]]},
        {"level":"secondary","points":[[0.17,0.55],[0.34,0.60],[0.52,0.63],[0.75,0.72]]},
        {"level":"secondary","points":[[0.19,0.71],[0.38,0.73],[0.59,0.69],[0.80,0.49]]}
      ]
    }$json$::jsonb
),
(
    '320500', 1, 120.5853, 31.2989, 119.92, 30.76, 121.05, 31.78,
    $json${
      "outline":[[0.09,0.25],[0.21,0.11],[0.46,0.07],[0.70,0.13],[0.88,0.28],[0.93,0.49],[0.84,0.70],[0.65,0.86],[0.41,0.92],[0.19,0.80],[0.07,0.59],[0.09,0.25]],
      "waterways":[[[0.09,0.31],[0.25,0.35],[0.40,0.42],[0.55,0.40],[0.73,0.33],[0.88,0.37]],[[0.18,0.69],[0.32,0.60],[0.48,0.58],[0.64,0.64],[0.80,0.73]],[[0.48,0.10],[0.47,0.27],[0.50,0.43],[0.47,0.59],[0.44,0.83]]],
      "roads":[
        {"level":"primary","points":[[0.12,0.52],[0.30,0.48],[0.49,0.49],[0.68,0.53],[0.88,0.48]]},
        {"level":"primary","points":[[0.22,0.18],[0.33,0.33],[0.46,0.48],[0.60,0.65],[0.73,0.82]]},
        {"level":"primary","points":[[0.15,0.73],[0.31,0.64],[0.48,0.58],[0.66,0.44],[0.82,0.27]]},
        {"level":"secondary","points":[[0.25,0.13],[0.26,0.32],[0.29,0.50],[0.28,0.77]]},
        {"level":"secondary","points":[[0.40,0.09],[0.39,0.27],[0.42,0.47],[0.39,0.70],[0.42,0.88]]},
        {"level":"secondary","points":[[0.58,0.10],[0.57,0.30],[0.58,0.50],[0.61,0.73],[0.62,0.86]]},
        {"level":"secondary","points":[[0.75,0.18],[0.70,0.35],[0.70,0.55],[0.77,0.74]]},
        {"level":"secondary","points":[[0.12,0.40],[0.31,0.39],[0.51,0.35],[0.72,0.29]]},
        {"level":"secondary","points":[[0.17,0.61],[0.35,0.69],[0.55,0.71],[0.78,0.64]]}
      ]
    }$json$::jsonb
),
(
    '320100', 1, 118.7969, 32.0603, 118.35, 31.55, 119.23, 32.61,
    $json${
      "outline":[[0.16,0.10],[0.40,0.06],[0.65,0.13],[0.84,0.29],[0.91,0.52],[0.81,0.74],[0.60,0.90],[0.35,0.88],[0.15,0.74],[0.07,0.49],[0.10,0.27],[0.16,0.10]],
      "waterways":[[[0.08,0.70],[0.23,0.62],[0.37,0.53],[0.51,0.43],[0.66,0.35],[0.85,0.31]],[[0.56,0.12],[0.54,0.29],[0.51,0.43],[0.55,0.60],[0.64,0.82]]],
      "roads":[
        {"level":"primary","points":[[0.12,0.56],[0.29,0.50],[0.47,0.47],[0.65,0.48],[0.85,0.54]]},
        {"level":"primary","points":[[0.21,0.17],[0.33,0.31],[0.46,0.46],[0.58,0.64],[0.69,0.84]]},
        {"level":"primary","points":[[0.13,0.35],[0.31,0.37],[0.50,0.34],[0.70,0.27],[0.83,0.22]]},
        {"level":"primary","points":[[0.18,0.76],[0.36,0.67],[0.54,0.61],[0.76,0.66]]},
        {"level":"secondary","points":[[0.27,0.11],[0.28,0.30],[0.30,0.50],[0.29,0.75]]},
        {"level":"secondary","points":[[0.43,0.08],[0.42,0.28],[0.45,0.47],[0.44,0.69],[0.48,0.87]]},
        {"level":"secondary","points":[[0.62,0.13],[0.59,0.30],[0.60,0.49],[0.64,0.72]]},
        {"level":"secondary","points":[[0.76,0.22],[0.71,0.39],[0.71,0.57],[0.77,0.73]]},
        {"level":"secondary","points":[[0.12,0.64],[0.31,0.60],[0.50,0.64],[0.73,0.76]]}
      ]
    }$json$::jsonb
)
ON CONFLICT (city_code) DO UPDATE SET
    version = EXCLUDED.version,
    center_longitude = EXCLUDED.center_longitude,
    center_latitude = EXCLUDED.center_latitude,
    min_longitude = EXCLUDED.min_longitude,
    min_latitude = EXCLUDED.min_latitude,
    max_longitude = EXCLUDED.max_longitude,
    max_latitude = EXCLUDED.max_latitude,
    geometry = EXCLUDED.geometry,
    updated_at = NOW();

COMMIT;
