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
    creator_id INT REFERENCES users(id),
    remover_id INT REFERENCES users(id),
    target_id INT REFERENCES users(id),
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
    makerspace_id INT REFERENCES makerspaces(id) ON DELETE CASCADE,
    creator_id INT REFERENCES users(id),
    target_id INT REFERENCES users(id),
    create_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    reason TEXT NOT NULL DEFAULT ''
);

CREATE TABLE zones (
    id SERIAL PRIMARY KEY,
    makerspace_id INT REFERENCES makerspaces(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    hidden BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE default_hours (
    makerspace_id INT REFERENCES makerspace_id(id) ON DELETE CASCADE,
    day_of_week INT NOT NULL CHECK (day_of_week >= 0 AND day_of_week < 7),
    open_time TIME WITH TIME ZONE,
    close_time TIME WITH TIME ZONE,
    closed BOOLEAN NOT NULL DEFAULT TRUE
);

CREATE TABLE special_hours (
    makerspace_id INT REFERENCES makerspace_id(id) ON DELETE CASCADE,
    special_date DATE NOT NULL,
    open_time TIME WITH TIME ZONE,
    close_time TIME WITH TIME ZONE,
    closed BOOLEAN NOT NULL DEFAULT TRUE
);

-- +goose Down
DROP TABLE special_hours;
DROP TABLE default_hours;
DROP TABLE zones;
DROP TABLE restrictions;
DROP TABLE makerspaces;
DROP TABLE images;
DROP TABLE holds;
DROP TABLE users;
