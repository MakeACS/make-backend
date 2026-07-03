-- +goose Up
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL,
    last_name TEXT NOT NULL,
    pronouns TEXT NOT NULL DEFAULT '',
    join_date TIMESTAMP WITH TIME ZONE NOT NULL,
    setup_complete BOOLEAN NOT NULL DEFAULT FALSE,
    archived BOOLEAN,
    notes TEXT NOT NULL DEFAULT '',
    admin BOOLEAN NOT NULL DEFAULT FALSE,
    force_archive BOOLEAN,
    card_tag TEXT NOT NULL DEFAULT ''
);

CREATE TABLE holds (
    id SERIAL PRIMARY KEY,
    creator_id INT NOT NULL REFERENCES users(id),
    remover_id INT REFERENCES users(id),
    target_id INT NOT NULL REFERENCES users(id),
    reason TEXT NOT NULL DEFAULT '',
    create_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    remove_date TIMESTAMP WITH TIME ZONE
);

CREATE TABLE images (
    id SERIAL PRIMARY KEY,
    identifier TEXT NOT NULL
);

CREATE TABLE makerspaces (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    subtitle TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    docs_url TEXT NOT NULL DEFAULT '',
    image_id INT REFERENCES images(id) ON DELETE SET NULL,
    hidden BOOLEAN NOT NULL
);

CREATE TABLE restrictions (
    id SERIAL PRIMARY KEY,
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    creator_id INT NOT NULL REFERENCES users(id),
    target_id INT NOT NULL REFERENCES users(id),
    create_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reason TEXT NOT NULL DEFAULT ''
);

CREATE TABLE zones (
    id SERIAL PRIMARY KEY,
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    hidden BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE default_hours (
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week < 7),
    open_time TIME WITH TIME ZONE,
    close_time TIME WITH TIME ZONE,
    closed BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (makerspace_id, day_of_week)
);

CREATE TABLE special_hours (
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    special_date DATE NOT NULL,
    open_time TIME WITH TIME ZONE,
    close_time TIME WITH TIME ZONE,
    closed BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (makerspace_id, special_date)
);

CREATE TABLE announcements (
    id SERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL DEFAULT '',
    makerspace_id INT REFERENCES makerspaces(id) ON DELETE CASCADE
);

CREATE TABLE managers (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, makerspace_id)
);

CREATE TABLE staff (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, makerspace_id)
);

CREATE TABLE equipment (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    sub_name TEXT NOT NULL DEFAULT '',
    zone_id INT NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    hidden BOOLEAN NOT NULL DEFAULT FALSE,
    image_id INT REFERENCES images(id) ON DELETE SET NULL,
    sop_url TEXT NOT NULL DEFAULT '',
    sign_off_url TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    reservable BOOLEAN NOT NULL DEFAULT FALSE,
    reservation_only BOOLEAN NOT NULL DEFAULT FALSE,
    needs_welcome BOOLEAN NOT NULL DEFAULT TRUE,
    requires_in_person BOOLEAN NOT NULL DEFAULT TRUE,
    requires_trainer BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE trainers (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    equipment_id INT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, equipment_id)
);

CREATE TABLE welcome_taps (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE SET NULL,
    tap_time TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    makerspace_id INT REFERENCES makerspaces(id) ON DELETE SET NULL
);

CREATE TABLE trainings (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    questions JSONB NOT NULL,
    makerspace_id INT REFERENCES makerspaces(id) ON DELETE CASCADE
);

CREATE TABLE equipment_trainings (
    equipment_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    training_id  INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE
);

CREATE TABLE zone_trainings (
    zone_id INT NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    training_id INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE zone_trainings;
DROP TABLE equipment_trainings;
DROP TABLE trainings;
DROP TABLE welcome_taps;
DROP TABLE trainers;
DROP TABLE equipment;
DROP TABLE staff;
DROP TABLE managers;
DROP TABLE announcements;
DROP TABLE special_hours;
DROP TABLE default_hours;
DROP TABLE zones;
DROP TABLE restrictions;
DROP TABLE makerspaces;
DROP TABLE images;
DROP TABLE holds;
DROP TABLE users;
