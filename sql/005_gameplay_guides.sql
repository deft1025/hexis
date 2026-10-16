CREATE TABLE IF NOT EXISTS gameplay_guides (
    patch_id BIGINT UNSIGNED NOT NULL,
    guide_key VARCHAR(100) NOT NULL,
    title VARCHAR(200) NOT NULL,
    category VARCHAR(20) NOT NULL,
    content JSON NOT NULL,
    source_id BIGINT UNSIGNED NOT NULL,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    PRIMARY KEY (patch_id, guide_key),
    CONSTRAINT fk_guides_patch FOREIGN KEY (patch_id) REFERENCES patches(id),
    CONSTRAINT fk_guides_source FOREIGN KEY (source_id) REFERENCES data_sources(id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
