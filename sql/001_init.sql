CREATE DATABASE IF NOT EXISTS hexis
    CHARACTER SET utf8mb4
    COLLATE utf8mb4_unicode_ci;

USE hexis;

CREATE TABLE patches (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    version VARCHAR(20) NOT NULL,
    status ENUM('draft', 'collecting', 'published', 'archived') NOT NULL DEFAULT 'draft',
    released_at DATETIME NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_patches_version (version)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE champions (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    riot_id VARCHAR(64) NOT NULL,
    name_zh VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NULL,
    icon_url VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_champions_riot_id (riot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE augments (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    riot_id VARCHAR(128) NOT NULL,
    name_zh VARCHAR(200) NOT NULL,
    name_en VARCHAR(200) NULL,
    description TEXT NULL,
    rarity VARCHAR(30) NULL,
    icon_url VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_augments_riot_id (riot_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE champion_patch_stats (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    patch_id BIGINT UNSIGNED NOT NULL,
    champion_id BIGINT UNSIGNED NOT NULL,
    mode VARCHAR(30) NOT NULL DEFAULT 'hexis',
    sample_count INT UNSIGNED NOT NULL DEFAULT 0,
    pick_rate DECIMAL(8,5) NULL,
    win_rate DECIMAL(8,5) NULL,
    tier VARCHAR(10) NULL,
    source_id BIGINT UNSIGNED NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_champion_patch_mode (patch_id, champion_id, mode),
    KEY idx_champion_stats_patch (patch_id),
    CONSTRAINT fk_champion_stats_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_champion_stats_champion FOREIGN KEY (champion_id) REFERENCES champions(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE augment_patch_stats (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    patch_id BIGINT UNSIGNED NOT NULL,
    augment_id BIGINT UNSIGNED NOT NULL,
    sample_count INT UNSIGNED NOT NULL DEFAULT 0,
    pick_rate DECIMAL(8,5) NULL,
    win_rate DECIMAL(8,5) NULL,
    tier VARCHAR(10) NULL,
    source_id BIGINT UNSIGNED NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (id),
    UNIQUE KEY uk_augment_patch (patch_id, augment_id),
    KEY idx_augment_stats_patch (patch_id),
    CONSTRAINT fk_augment_stats_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_augment_stats_augment FOREIGN KEY (augment_id) REFERENCES augments(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE data_sources (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    name VARCHAR(100) NOT NULL,
    url VARCHAR(500) NULL,
    license VARCHAR(100) NULL,
    attribution VARCHAR(500) NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE import_jobs (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    source_id BIGINT UNSIGNED NOT NULL,
    patch_id BIGINT UNSIGNED NOT NULL,
    kind VARCHAR(30) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_hash CHAR(64) NOT NULL,
    file_size BIGINT UNSIGNED NOT NULL DEFAULT 0,
    status ENUM('running', 'success', 'failed') NOT NULL,
    error_message TEXT NULL,
    started_at DATETIME NULL,
    finished_at DATETIME NULL,
    PRIMARY KEY (id),
    KEY idx_import_jobs_patch (patch_id),
    CONSTRAINT fk_import_jobs_source FOREIGN KEY (source_id) REFERENCES data_sources(id),
    CONSTRAINT fk_import_jobs_patch FOREIGN KEY (patch_id) REFERENCES patches(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
