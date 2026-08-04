CREATE TABLE workouts (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    external_id TEXT NOT NULL UNIQUE,
    activity TEXT NOT NULL,
    activity_group TEXT NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    duration_s DOUBLE PRECISION NOT NULL,
    distance_m DOUBLE PRECISION NOT NULL,
    active_s DOUBLE PRECISION NOT NULL,
    elapsed_s DOUBLE PRECISION NOT NULL,
    paused_s DOUBLE PRECISION NOT NULL,
    calories DOUBLE PRECISION,
    run_distance_m DOUBLE PRECISION,
    walk_distance_m DOUBLE PRECISION,
    elevation_gain_m DOUBLE PRECISION,
    elevation_loss_m DOUBLE PRECISION,
    gait_run_distance_m DOUBLE PRECISION,
    gait_walk_distance_m DOUBLE PRECISION,
    gait_overall_pace_s_per_km DOUBLE PRECISION,
    gait_overall_cadence_spm DOUBLE PRECISION,
    gait_run_pace_s_per_km DOUBLE PRECISION,
    gait_run_cadence_spm DOUBLE PRECISION,
    gait_walk_pace_s_per_km DOUBLE PRECISION,
    gait_walk_cadence_spm DOUBLE PRECISION,
    weather_temp_c DOUBLE PRECISION,
    weather_apparent_temp_c DOUBLE PRECISION,
    weather_humidity_pct DOUBLE PRECISION,
    weather_code INTEGER,
    weather_label TEXT,
    weather_precip_mm DOUBLE PRECISION,
    weather_snowfall_cm DOUBLE PRECISION,
    weather_cloud_cover_pct DOUBLE PRECISION,
    weather_visibility_m DOUBLE PRECISION
);

CREATE INDEX idx_workouts_start_time ON workouts(start_time);

CREATE TABLE workout_segments (
    id INTEGER GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    workout_id INTEGER NOT NULL REFERENCES workouts(id) ON DELETE CASCADE,
    segment_index INTEGER NOT NULL,
    distance_m DOUBLE PRECISION NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    elapsed_s DOUBLE PRECISION NOT NULL,
    pace_s_per_km DOUBLE PRECISION NOT NULL,
    speed_kmh DOUBLE PRECISION NOT NULL,
    gap_elapsed_s DOUBLE PRECISION NOT NULL,
    gap_pace_s_per_km DOUBLE PRECISION NOT NULL,
    is_pb BOOLEAN NOT NULL,
    previous_best_s DOUBLE PRECISION,
    best_s DOUBLE PRECISION,
    rank INTEGER NOT NULL,
    UNIQUE (workout_id, segment_index)
);

CREATE INDEX idx_workout_segments_workout_id ON workout_segments(workout_id);
