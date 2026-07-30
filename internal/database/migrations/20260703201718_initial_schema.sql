-- +goose Up

CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),  
    makerspace_id INT NULL DEFAULT NULL,
    plain_string TEXT NOT NULL ,
    format_string TEXT NOT NULL,
    message_type TEXT NOT NULL,
    data JSONB NULL
);

CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    first_name TEXT NOT NULL DEFAULT '',
    last_name TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL UNIQUE,
    pronouns TEXT NOT NULL DEFAULT '',
    join_date TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    setup_complete BOOLEAN NOT NULL DEFAULT FALSE,
    archived BOOLEAN NOT NULL DEFAULT FALSE,
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
    identifier TEXT NOT NULL UNIQUE
);



CREATE TABLE groups (
    id SERIAL PRIMARY KEY,
    name TEXT UNIQUE NOT NULL,
    manager_id INT REFERENCES groups(id),
    description TEXT NOT NULL DEFAULT '',
    UNIQUE (name, manager_id)
);
INSERT INTO groups (name, description) 
VALUES ('admin', 'The root of all groups');



CREATE TABLE group_direct_membership(
    group_id INT NOT NULL REFERENCES groups(id) ON DELETE CASCADE,
    user_id INT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    view_permission INT NOT NULL DEFAULT 0 CHECK (view_permission in (0, 1, 2)),
    PRIMARY KEY (group_id, user_id)
);

CREATE TABLE group_direct_subgroups (
    group_id INT REFERENCES groups(id),
    subgroup_id INT REFERENCES groups(id),
    view_permission INT NOT NULL DEFAULT 0 CHECK (view_permission in (0, 1, 2)),
    PRIMARY KEY (group_id, subgroup_id)
);


-- table of
-- manager_id, managed_id, depth
-- admin, staff_management, 1
-- admin, student_groups, 1
-- student_groups, club1, 1
-- admin, club1, 2

-- does not take into account subgroups
-- if 
--   A super:subgroups B 
--   A manages C
--   C manages D
-- this view only shows A manages C, C manages D, A manages C
-- the group_management view includes B manages C, B manages D

create view group_noncombination_management as ( 
	WITH RECURSIVE managees AS (
		-- Anchor: get all groups and their managers
	    SELECT g.manager_id as manager_group_id, g.id as group_id, 1 AS depth 
	    FROM "groups" g
	    UNION ALL
	    -- Recursive: Joins my id and existing managers to find existing groups connection up the tree
	    SELECT g2.manager_id as manager_group_id, m.group_id as group_id, m.depth + 1
	    FROM groups g2
	    JOIN managees m ON g2.id = m.manager_group_id
	    where g2.id <> g2.manager_id and depth < 1000
	)
	SELECT manager_group_id, group_id, min(depth) as depth FROM managees group by (manager_group_id, group_id) 
);

create view group_subgroups as (
WITH RECURSIVE subgroups AS (
		-- Anchor: get all subgroups groups and their direct supergroups
	    SELECT g.group_id as supergroup_id, g.subgroup_id as subgroup_id, view_permission, 1 AS depth 
	    FROM "group_direct_subgroups" g
	    UNION ALL
	    -- Recursive: Joins to find existing groups connection up the tree
	    -- geometrically working up the tree
	    SELECT g2.group_id as supergroup_id, s.subgroup_id as subgroup_id, g2.view_permission, s.depth + 1
	    FROM "group_direct_subgroups" g2
	    JOIN subgroups s ON g2.subgroup_id  = s.supergroup_id 
	    where g2.group_id <> g2.subgroup_id and depth < 1000
	)
SELECT distinct supergroup_id, subgroup_id, view_permission, depth as depth 
FROM subgroups 
);

-- takes into account subgroups
-- if 
--   A super:subgroups B 
--   A manages C
--   C manages D
-- this view only shows A manages C, C manages D, A manages C, B manages C, B manages D
-- the group_noncombination_management view only shows A:C, C:D, A:C as it is unaware of subgroups

create view group_management as ( 
	with all_manage_links as (
		select gnm.manager_group_id as manager_group_id, gnm.group_id,0 as subgroup_depth, gnm.depth as manage_depth from group_noncombination_management gnm 
	union
		select 
			gs.subgroup_id as manager_group_id, 
			gnm.group_id as group_id, 
			gs.depth as subgroup_depth, 
			gnm.depth as manage_depth
		from group_subgroups gs 
		left join group_noncombination_management gnm 
		on gs.supergroup_id  = gnm.manager_group_id
		where gnm.group_id is not null
	)
	select manager_group_id , group_id, min(subgroup_depth) as subgroup_depth, min(manage_depth) as manage_depth  from all_manage_links
	group by manager_group_id , group_id 
);

create view group_membership as ( 
	select group_id, user_id, gdm.view_permission  from group_direct_membership gdm 
	union
	select distinct gs.supergroup_id, gdm.user_id, gs.view_permission  
	from group_subgroups gs 
	left join group_direct_membership gdm 
	on gdm.group_id  = gs.subgroup_id 
	where user_id is not null
);



CREATE TABLE anonymous_groups(
    id SERIAL PRIMARY KEY
);
CREATE TABLE anonymous_group_subgroups(
    anonymous_id INT REFERENCES anonymous_groups(id) ON DELETE CASCADE,
    group_id INT REFERENCES groups(id) ON DELETE CASCADE, 
    PRIMARY KEY (anonymous_id, group_id)
);



CREATE TABLE makerspaces (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    subtitle TEXT NOT NULL DEFAULT '',
    location TEXT NOT NULL DEFAULT '',
    description TEXT NOT NULL DEFAULT '',
    docs_url TEXT NOT NULL DEFAULT '',
    image_id INT REFERENCES images(id) ON DELETE SET NULL,
    hidden BOOLEAN NOT NULL,
    timezone TEXT NOT NULL DEFAULT 'America/New_York',
    -- agroup of users who can manage this space
    management_agroup_id INT NOT NULL REFERENCES anonymous_groups(id) ON DELETE CASCADE,
    -- agroup of users who can site-set equipment state
    can_change_equipment_state_agroup_id INT NOT NULL REFERENCES anonymous_groups(id) ON DELETE CASCADE 
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
    open_time TIME,
    close_time TIME,
    closed BOOLEAN NOT NULL DEFAULT TRUE,
    PRIMARY KEY (makerspace_id, day_of_week)
);

CREATE TABLE special_hours (
    makerspace_id INT NOT NULL REFERENCES makerspaces(id) ON DELETE CASCADE,
    special_date DATE NOT NULL,
    open_time TIME,
    close_time TIME,
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
    blocks JSONB NOT NULL,
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
    temp_duration INT NOT NULL DEFAULT 100,
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
    temp_duration INT NOT NULL DEFAULT 100
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



-- +goose Down
DROP TABLE IF EXISTS groups;
DROP TABLE IF EXISTS group_direct_membership;
DROP TABLE IF EXISTS group_direct_subgroups;


DROP TABLE IF EXISTS anonymous_groups;
DROP TABLE IF EXISTS anonymous_group_subgroups;

DROP TABLE IF EXISTS custom_links;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS equipment_instances;
DROP TABLE IF EXISTS welcome_devices;
DROP TABLE IF EXISTS dispensers;
DROP TYPE IF EXISTS DISPENSER_ERROR;
DROP TABLE IF EXISTS access_channels;
DROP TYPE IF EXISTS ACCESS_CHANNEL_STATE;
DROP TABLE IF EXISTS access_devices;
DROP TABLE IF EXISTS devices;
DROP TABLE IF EXISTS organization_members;
DROP TABLE IF EXISTS organizations;
DROP TABLE IF EXISTS passed_trainings;
DROP TABLE IF EXISTS training_holds;
DROP TABLE IF EXISTS makerspace_trainings;
DROP TABLE IF EXISTS zone_trainings;
DROP TABLE IF EXISTS equipment_trainings;
DROP TABLE IF EXISTS trainings;
DROP TABLE IF EXISTS welcome_taps;
DROP TABLE IF EXISTS trainers;
DROP TABLE IF EXISTS equipment;
DROP TABLE IF EXISTS staff;
DROP TABLE IF EXISTS managers;
DROP TABLE IF EXISTS announcements;
DROP TABLE IF EXISTS special_hours;
DROP TABLE IF EXISTS default_hours;
DROP TABLE IF EXISTS zones;
DROP TABLE IF EXISTS restrictions;
DROP TABLE IF EXISTS makerspaces;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS holds;
DROP TABLE IF EXISTS users;
