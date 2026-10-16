CREATE TABLE IF NOT EXISTS champion_patch_skills (
    patch_id BIGINT UNSIGNED NOT NULL,
    champion_id BIGINT UNSIGNED NOT NULL,
    data_version VARCHAR(30) NOT NULL,
    title VARCHAR(100) NOT NULL,
    roles JSON NOT NULL,
    skills JSON NOT NULL,
    source_url VARCHAR(500) NOT NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    PRIMARY KEY (patch_id, champion_id),
    CONSTRAINT fk_skills_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_skills_champion FOREIGN KEY (champion_id) REFERENCES champions(id),
    CONSTRAINT fk_skills_source FOREIGN KEY (source_id) REFERENCES data_sources(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
