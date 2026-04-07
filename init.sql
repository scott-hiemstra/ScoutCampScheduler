CREATE DATABASE IF NOT EXISTS camp;
USE camp;

-- Schema
CREATE TABLE units (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE
) ENGINE=InnoDB;

CREATE TABLE scouts (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    unit_id INT NOT NULL,
    fill_schedule BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (unit_id) REFERENCES units(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE program_areas (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    color VARCHAR(7) NOT NULL DEFAULT '#6c757d'
) ENGINE=InnoDB;

CREATE TABLE time_blocks (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(10) NOT NULL UNIQUE,
    start_time VARCHAR(10) NOT NULL,
    end_time VARCHAR(10) NOT NULL
) ENGINE=InnoDB;

CREATE TABLE activities (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    program_area_id INT NOT NULL,
    time_block_id INT NOT NULL,
    capacity INT NOT NULL,
    half_week ENUM('first','second','full') NOT NULL,
    has_prerequisites BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (program_area_id) REFERENCES program_areas(id),
    FOREIGN KEY (time_block_id) REFERENCES time_blocks(id)
) ENGINE=InnoDB;

CREATE TABLE scout_preferences (
    id INT AUTO_INCREMENT PRIMARY KEY,
    scout_id INT NOT NULL,
    activity_name VARCHAR(100) NOT NULL,
    priority INT NOT NULL DEFAULT 0,
    FOREIGN KEY (scout_id) REFERENCES scouts(id) ON DELETE CASCADE
) ENGINE=InnoDB;

CREATE TABLE assignments (
    id INT AUTO_INCREMENT PRIMARY KEY,
    scout_id INT NOT NULL,
    activity_id INT NOT NULL,
    locked BOOLEAN NOT NULL DEFAULT FALSE,
    FOREIGN KEY (scout_id) REFERENCES scouts(id) ON DELETE CASCADE,
    FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE,
    UNIQUE KEY unique_assignment (scout_id, activity_id)
) ENGINE=InnoDB;

-- Time blocks
INSERT INTO time_blocks (name, start_time, end_time) VALUES
('A', '9:00', '10:30'),
('B', '10:30', '12:00'),
('AB', '9:00', '12:00'),
('C', '2:00', '3:30'),
('D', '3:30', '5:00'),
('CD', '2:00', '5:00'),
('ABCD', '9:00', '5:00');

-- Program areas (colors inspired by the PDF)
INSERT INTO program_areas (name, color) VALUES
('Aquatics', '#0077b6'),
('Eco-STEM', '#2d6a4f'),
('Handicraft', '#e07b00'),
('Scout Scholar', '#5e35b1'),
('Scoutcraft', '#c62828'),
('COPE', '#546e7a'),
('Rifle & Target', '#6d4c41'),
('TFC', '#1565c0'),
('Trek', '#558b2f');

-- ============================================================
-- ACTIVITIES SEED DATA from 2026 Scouts BSA Activity Schedule
-- ============================================================
-- Program area IDs: 1=Aquatics, 2=Eco-STEM, 3=Handicraft,
--   4=Scout Scholar, 5=Scoutcraft, 6=COPE, 7=Rifle & Target,
--   8=TFC, 9=Trek
-- Time block IDs: 1=A, 2=B, 3=AB, 4=C, 5=D, 6=CD, 7=ABCD

-- ==================
-- AQUATICS (area 1)
-- ==================
-- Block A  (swimming/lifesaving run full week; others are per-half)
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Swimming', 1, 1, 10, 'full', TRUE),
('Lifesaving', 1, 1, 10, 'full', TRUE),
('Motorboating + Rowing', 1, 1, 6, 'first', FALSE),
('Motorboating + Rowing', 1, 1, 6, 'second', FALSE),
('Kayaking', 1, 1, 12, 'first', FALSE),
('Kayaking', 1, 1, 12, 'second', FALSE),
('Canoeing', 1, 1, 12, 'first', FALSE),
('Canoeing', 1, 1, 12, 'second', FALSE);

-- Block B
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Swimming', 1, 2, 10, 'full', TRUE),
('Lifesaving', 1, 2, 10, 'full', TRUE),
('Small Boat Sailing', 1, 2, 12, 'first', FALSE),
('Small Boat Sailing', 1, 2, 12, 'second', FALSE),
('Kayaking', 1, 2, 12, 'first', FALSE),
('Kayaking', 1, 2, 12, 'second', FALSE),
('Canoeing', 1, 2, 12, 'first', FALSE),
('Canoeing', 1, 2, 12, 'second', FALSE);

-- Block AB
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Watersports', 1, 3, 5, 'first', FALSE),
('Watersports', 1, 3, 5, 'second', FALSE);

-- Block C
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Swimming', 1, 4, 10, 'full', TRUE),
('Lifesaving', 1, 4, 10, 'full', TRUE),
('Small Boat Sailing', 1, 4, 12, 'first', FALSE),
('Small Boat Sailing', 1, 4, 12, 'second', FALSE),
('Kayaking', 1, 4, 12, 'first', FALSE),
('Kayaking', 1, 4, 12, 'second', FALSE),
('Canoeing', 1, 4, 12, 'first', FALSE),
('Canoeing', 1, 4, 12, 'second', FALSE);

-- Block CD
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Watersports', 1, 6, 5, 'first', FALSE),
('Watersports', 1, 6, 5, 'second', FALSE);

-- ==================
-- ECO-STEM (area 2)
-- ==================
-- Block A
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Space Exploration', 2, 1, 16, 'first', FALSE),
('Reptile and Amphibian', 2, 1, 16, 'first', FALSE),
('Geology', 2, 1, 16, 'first', FALSE),
('Cybersecurity', 2, 1, 12, 'second', FALSE),
('Chemistry', 2, 1, 16, 'second', FALSE),
('Nature', 2, 1, 16, 'second', FALSE);

-- Block B
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Space Exploration', 2, 2, 16, 'first', FALSE),
('Reptile and Amphibian', 2, 2, 16, 'first', FALSE),
('Geology', 2, 2, 16, 'first', FALSE),
('Cybersecurity', 2, 2, 12, 'second', FALSE),
('Chemistry', 2, 2, 16, 'second', FALSE),
('Insect Study', 2, 2, 16, 'second', FALSE);

-- Block AB
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Environmental Science', 2, 3, 16, 'full', TRUE);

-- Block C
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Astronomy', 2, 4, 16, 'first', FALSE),
('Forestry', 2, 4, 16, 'first', FALSE),
('Engineering', 2, 4, 16, 'first', FALSE),
('Astronomy', 2, 4, 16, 'second', FALSE),
('Soil & Water Conservation', 2, 4, 16, 'second', FALSE),
('Engineering', 2, 4, 16, 'second', FALSE);

-- Block D
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Artificial Intelligence', 2, 5, 16, 'first', FALSE),
('Mammal Study', 2, 5, 16, 'first', FALSE),
('Artificial Intelligence', 2, 5, 16, 'second', FALSE),
('Mammal Study', 2, 5, 16, 'second', FALSE);

-- Block CD
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Environmental Science', 2, 6, 16, 'full', TRUE);

-- =====================
-- HANDICRAFT (area 3)
-- =====================
-- Block A
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Metalwork', 3, 1, 12, 'first', FALSE),
('Basketry', 3, 1, 12, 'first', FALSE),
('Graphic Arts', 3, 1, 12, 'first', FALSE),
('Woodcarving', 3, 1, 12, 'second', FALSE),
('Leatherwork', 3, 1, 12, 'second', FALSE),
('Graphic Arts', 3, 1, 12, 'second', FALSE);

-- Block B
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Metalwork', 3, 2, 12, 'first', FALSE),
('Basketry', 3, 2, 12, 'first', FALSE),
('Game Design', 3, 2, 12, 'first', FALSE),
('Woodcarving', 3, 2, 12, 'second', FALSE),
('Leatherwork', 3, 2, 12, 'second', FALSE),
('Game Design', 3, 2, 12, 'second', FALSE);

-- Block C
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Metalwork', 3, 4, 12, 'first', FALSE),
('Basketry', 3, 4, 12, 'first', FALSE),
('Sculpture', 3, 4, 12, 'first', FALSE),
('Woodcarving', 3, 4, 12, 'second', FALSE),
('Leatherwork', 3, 4, 12, 'second', FALSE),
('Sculpture', 3, 4, 12, 'second', FALSE);

-- Block D
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Pulp and Paper', 3, 5, 12, 'first', FALSE),
('Fingerprinting', 3, 5, 24, 'first', FALSE),
('Pulp and Paper', 3, 5, 12, 'second', FALSE),
('Fingerprinting', 3, 5, 24, 'second', FALSE);

-- ========================
-- SCOUT SCHOLAR (area 4)
-- ========================
-- Block A
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Citizenship in the Nation', 4, 1, 16, 'first', TRUE),
('Emergency Preparedness', 4, 1, 16, 'first', TRUE),
('Chess', 4, 1, 16, 'first', FALSE),
('Citizenship in the World', 4, 1, 16, 'second', TRUE),
('Emergency Preparedness', 4, 1, 16, 'second', TRUE),
('Law', 4, 1, 16, 'second', FALSE);

-- Block B
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Citizenship in the Nation', 4, 2, 16, 'first', TRUE),
('Emergency Preparedness', 4, 2, 16, 'first', TRUE),
('Chess', 4, 2, 16, 'first', FALSE),
('Citizenship in the World', 4, 2, 16, 'second', TRUE),
('Emergency Preparedness', 4, 2, 16, 'second', TRUE),
('Law', 4, 2, 16, 'second', FALSE);

-- Block AB
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('First Aid', 4, 3, 16, 'full', TRUE);

-- Block C
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Citizenship in the World', 4, 4, 16, 'first', TRUE),
('Chess', 4, 4, 16, 'first', FALSE),
('Search and Rescue', 4, 4, 16, 'first', FALSE),
('Citizenship in the Nation', 4, 4, 16, 'second', TRUE),
('Music', 4, 4, 16, 'second', FALSE),
('Search and Rescue', 4, 4, 16, 'second', FALSE);

-- Block CD
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Communications', 4, 6, 12, 'full', TRUE);

-- ======================
-- SCOUTCRAFT (area 5)
-- ======================
-- Block A
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Camping', 5, 1, 16, 'first', TRUE),
('Wilderness Survival', 5, 1, 16, 'first', FALSE),
('Fishing', 5, 1, 16, 'first', FALSE),
('Camping', 5, 1, 16, 'second', TRUE),
('Geocaching', 5, 1, 16, 'second', FALSE),
('Fishing', 5, 1, 16, 'second', FALSE);

-- Block B
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Cooking', 5, 2, 12, 'first', TRUE),
('Orienteering', 5, 2, 16, 'first', FALSE),
('Fishing', 5, 2, 16, 'first', FALSE),
('Cooking', 5, 2, 12, 'second', TRUE),
('Wilderness Survival', 5, 2, 16, 'second', FALSE),
('Fishing', 5, 2, 16, 'second', FALSE);

-- Block AB
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Pioneering', 5, 3, 16, 'full', FALSE);

-- Block C
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Camping', 5, 4, 16, 'first', TRUE),
('Wilderness Survival', 5, 4, 16, 'first', FALSE),
('Cooking', 5, 4, 12, 'first', TRUE),
('Camping', 5, 4, 16, 'second', TRUE),
('Fire Safety', 5, 4, 16, 'second', FALSE),
('Geocaching', 5, 4, 16, 'second', FALSE);

-- Block CD
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Pioneering', 5, 6, 16, 'full', FALSE);

-- ===============
-- COPE (area 6)
-- ===============
-- Block AB
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Climbing', 6, 3, 6, 'first', FALSE),
('C.O.P.E.', 6, 3, 6, 'second', FALSE);

-- Block CD
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Climbing', 6, 6, 6, 'first', FALSE),
('C.O.P.E.', 6, 6, 6, 'second', FALSE);

-- ===========================
-- RIFLE & TARGET (area 7)
-- ===========================
-- Block AB
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Shotgun Shooting', 7, 3, 6, 'first', FALSE),
('Archery', 7, 3, 12, 'first', FALSE),
('Rifle Shooting', 7, 3, 12, 'first', FALSE),
('Shotgun Shooting', 7, 3, 6, 'second', FALSE),
('Archery', 7, 3, 12, 'second', FALSE),
('Rifle Shooting', 7, 3, 12, 'second', FALSE);

-- Block CD
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Shotgun Shooting', 7, 6, 6, 'first', FALSE),
('Archery', 7, 6, 12, 'first', FALSE),
('Rifle Shooting', 7, 6, 12, 'first', FALSE),
('Shotgun Shooting', 7, 6, 6, 'second', FALSE),
('Archery', 7, 6, 12, 'second', FALSE),
('Rifle Shooting', 7, 6, 12, 'second', FALSE);

-- ==============
-- TFC (area 8)
-- ==============
INSERT INTO activities (name, program_area_id, time_block_id, capacity, half_week, has_prerequisites) VALUES
('Trail to First Class', 8, 7, 36, 'full', FALSE);
