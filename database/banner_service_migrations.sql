DROP TABLE IF EXISTS banner_tag;
DROP TABLE IF EXISTS versions;
DROP TABLE IF EXISTS banners;
DROP TABLE IF EXISTS features;
DROP TABLE IF EXISTS tags;

CREATE TABLE features (
                      id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                      name TEXT NOT NULL UNIQUE
);

CREATE TABLE tags (
                      id INTEGER PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
                      name TEXT NOT NULL UNIQUE
);

CREATE TABLE banners (
                      id SERIAL PRIMARY KEY,
                      created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                      feature_id INTEGER,
                      FOREIGN KEY (feature_id) REFERENCES features(id)
);

CREATE TABLE banner_tag (
                      banner_id INTEGER REFERENCES banners ON DELETE CASCADE,
                      tag_id INTEGER REFERENCES tags ON DELETE CASCADE,
                      PRIMARY KEY (banner_id, tag_id)
);

CREATE TABLE versions (
                      id SERIAL PRIMARY KEY,
                      banner_id INTEGER REFERENCES banners ON DELETE CASCADE,
                      is_active BOOLEAN DEFAULT TRUE,
                      data JSONB NOT NULL,
                      updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

INSERT INTO features (name) VALUES
                                ('Feature 1'),
                                ('Feature 2'),
                                ('Feature 3');

INSERT INTO tags (name) VALUES
                            ('Tag 1'),
                            ('Tag 2'),
                            ('Tag 3');

INSERT INTO banners (id, created_at, feature_id) VALUES
                                                                (1, NOW(), 1),
                                                                (2, NOW(), 2),
                                                                (3, NOW(), 3);

INSERT INTO banner_tag (banner_id, tag_id) VALUES
                                               (1, 1),
                                               (1, 2),
                                               (2, 2),
                                               (3, 3);

INSERT INTO versions (banner_id, data, is_active, updated_at) VALUES
                                                       (2, '{"content": "Banner 2"}', TRUE,NOW()),
                                                       (3, '{"content": "Banner 3"}', TRUE,NOW());

INSERT INTO versions (banner_id, data, is_active, updated_at) VALUES
                                                       (1, '{"content": "Banner 1 - Version 1"}', TRUE,NOW() - INTERVAL '3 days'),
                                                       (1, '{"content": "Banner 1 - Version 2"}', FALSE,NOW() - INTERVAL '2 days'),
                                                       (1, '{"content": "Banner 1 - Version 3"}', FALSE,NOW() - INTERVAL '1 day'),
                                                       (1, '{"content": "Banner 1 - Version 4"}', FALSE,NOW());
