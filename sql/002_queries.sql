USE hexis;

CREATE TABLE champion_patch_profiles (
    patch_id BIGINT UNSIGNED NOT NULL,
    champion_id BIGINT UNSIGNED NOT NULL,
    name_zh VARCHAR(100) NOT NULL,
    name_en VARCHAR(100) NULL,
    icon_url VARCHAR(500) NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (patch_id, champion_id),
    CONSTRAINT fk_champion_profiles_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_champion_profiles_champion FOREIGN KEY (champion_id) REFERENCES champions(id),
    CONSTRAINT fk_champion_profiles_source FOREIGN KEY (source_id) REFERENCES data_sources(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE augment_patch_profiles (
    patch_id BIGINT UNSIGNED NOT NULL,
    augment_id BIGINT UNSIGNED NOT NULL,
    name_zh VARCHAR(200) NOT NULL,
    name_en VARCHAR(200) NULL,
    description TEXT NULL,
    rarity VARCHAR(30) NULL,
    icon_url VARCHAR(500) NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (patch_id, augment_id),
    CONSTRAINT fk_augment_profiles_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_augment_profiles_augment FOREIGN KEY (augment_id) REFERENCES augments(id),
    CONSTRAINT fk_augment_profiles_source FOREIGN KEY (source_id) REFERENCES data_sources(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

CREATE TABLE champion_combo_stats (
    patch_id BIGINT UNSIGNED NOT NULL,
    champion_a_id BIGINT UNSIGNED NOT NULL,
    champion_b_id BIGINT UNSIGNED NOT NULL,
    sample_count INT UNSIGNED NOT NULL DEFAULT 0,
    pick_rate DECIMAL(8,5) NULL,
    win_rate DECIMAL(8,5) NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (patch_id, champion_a_id, champion_b_id),
    CONSTRAINT fk_combo_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_combo_champion_a FOREIGN KEY (champion_a_id) REFERENCES champions(id),
    CONSTRAINT fk_combo_champion_b FOREIGN KEY (champion_b_id) REFERENCES champions(id),
    CONSTRAINT fk_combo_source FOREIGN KEY (source_id) REFERENCES data_sources(id),
    CONSTRAINT chk_combo_order CHECK (champion_a_id < champion_b_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
