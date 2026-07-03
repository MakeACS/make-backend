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
    reservation_instructions TEXT NOT NULL DEFAULT '',
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
    equipment_id INT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    training_id  INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    PRIMARY KEY (equipment_id, training_id)
);

CREATE TABLE zone_trainings (
    zone_id INT NOT NULL REFERENCES zones(id) ON DELETE CASCADE,
    training_id INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    PRIMARY KEY (zone_id, training_id)
);

CREATE TABLE makerspace_trainings (
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    training_id INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    PRIMARY KEY (makerspace_id, training_id)
);

CREATE TABLE training_holds (
    id SERIAL PRIMARY KEY,
    training_id INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    expires TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE passed_trainings (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    training_id INT NOT NULL REFERENCES trainings(id) ON DELETE CASCADE,
    passed_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, training_id)
);

CREATE TABLE organizations (
    id SERIAL PRIMARY KEY,
    display_name TEXT NOT NULL,
    email TEXT NOT NULL,
    notes TEXT NOT NULL DEFAULT ''
);

CREATE TABLE organization_members (
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    organization_id INT NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, organization_id)
);

CREATE TABLE devices (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    sn TEXT NOT NULL,
    paired TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    hardware TEXT,
    firmware TEXT,
    target_firmware TEXT,
    key_cycle INT NOT NULL DEFAULT 0,
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE 
);

CREATE TABLE access_devices (
    device_id INT PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    channels INT NOT NULL DEFAULT 0,
    temp_duration INTERVAL NOT NULL DEFAULT '100 MILLISECONDS',
    current_card_tag TEXT NOT NULL DEFAULT '',
    last_status TIMESTAMP WITH TIME ZONE,
    session_start TIMESTAMP WITH TIME ZONE,
    flags JSONB,
    sealed_deployment JSONB,
    reported_deployment JSONB
);

CREATE TYPE ACCESS_CHANNEL_STATE AS ENUM ('IDLE', 'UNLOCKED', 'ALWAYS_ON', 'LOCKED_OUT', 'FAULT');

CREATE TABLE access_channels (
    id SERIAL PRIMARY KEY,
    device_id INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    channel_id INT NOT NULL,
    state ACCESS_CHANNEL_STATE,
    temp_duration INTERVAL NOT NULL DEFAULT '100 MILLISECONDS'
);

CREATE TYPE DISPENSER_ERROR AS ENUM ('CARD_STUCK', 'OUT_OF_CARDS');

CREATE TABLE dispensers (
    device_id INT PRIMARY KEY REFERENCES devices(id) ON DELETE CASCADE,
    cards_left INT NOT NULL DEFAULT 0,
    error DISPENSER_ERROR
);

CREATE TABLE welcome_devices (
    device_id INT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    PRIMARY KEY (device_id, makerspace_id)
);

CREATE TABLE equipment_instances (
    id SERIAL PRIMARY KEY,
    equipment_id INT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    name TEXT NOT NULL DEFAULT '',
    access_channel_id INT REFERENCES access_channels(id) ON DELETE SET NULL
);

CREATE TABLE reservations (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    organization_id INT REFERENCES organizations(id) ON DELETE CASCADE,
    equipment_id INT NOT NULL REFERENCES equipment(id) ON DELETE CASCADE,
    description TEXT NOT NULL DEFAULT '',
    start_time  TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    approved BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE custom_links (
    short_url TEXT PRIMARY KEY,
    long_url TEXT NOT NULL
);

CREATE TABLE temp_cards (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    card_tag TEXT NOT NULL,
    issued TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- +goose Down
DROP TABLE temp_cards;
DROP TABLE custom_links;
DROP TABLE reservations;
DROP TABLE equipment_instances;
DROP TABLE welcome_devices;
DROP TABLE dispensers;
DROP TYPE DISPENSER_ERROR;
DROP TABLE access_channels;
DROP TYPE ACCESS_CHANNEL_STATE;
DROP TABLE access_devices;
DROP TABLE devices;
DROP TABLE organization_members;
DROP TABLE organizations;
DROP TABLE passed_trainings;
DROP TABLE training_holds;
DROP TABLE makerspace_trainings;
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
