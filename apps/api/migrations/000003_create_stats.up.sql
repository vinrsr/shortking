CREATE TABLE stats (
    id SMALLINT PRIMARY KEY DEFAULT 1,
    qr_generations BIGINT NOT NULL DEFAULT 0,
    CONSTRAINT stats_singleton CHECK (id = 1)
);

INSERT INTO stats (id) VALUES (1);
