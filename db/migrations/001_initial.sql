-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT        UNIQUE NOT NULL,
  password_hash TEXT        NOT NULL,
  first_name    TEXT,
  last_name     TEXT,
  phone         TEXT,
  role          TEXT        NOT NULL DEFAULT 'customer',
  reminder_opt  BOOLEAN     NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE barbers (
  id        UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id   UUID    UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name      TEXT    NOT NULL,
  title     TEXT,
  bio       TEXT,
  num       TEXT,
  active    BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE barber_schedules (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  barber_id  UUID NOT NULL REFERENCES barbers(id) ON DELETE CASCADE,
  dow        INT  NOT NULL,
  open_time  TIME NOT NULL,
  close_time TIME NOT NULL,
  UNIQUE (barber_id, dow)
);

CREATE TABLE services (
  id            UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  num           TEXT,
  name          TEXT    NOT NULL,
  name_html     TEXT,
  description   TEXT,
  duration_mins INT     NOT NULL,
  price_pence   INT     NOT NULL,
  active        BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE products (
  id          UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  name        TEXT    NOT NULL,
  price_pence INT     NOT NULL,
  active      BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE bookings (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  reference    TEXT        UNIQUE NOT NULL,
  user_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
  barber_id    UUID        NOT NULL REFERENCES barbers(id),
  service_id   UUID        NOT NULL REFERENCES services(id),
  date         DATE        NOT NULL,
  time_start   TIME        NOT NULL,
  time_end     TIME        NOT NULL,
  status       TEXT        NOT NULL DEFAULT 'confirmed',
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  cancelled_at TIMESTAMPTZ
);

CREATE TABLE booking_products (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id  UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  product_id  UUID NOT NULL REFERENCES products(id),
  qty         INT  NOT NULL,
  price_pence INT  NOT NULL
);

CREATE TABLE refresh_tokens (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT        NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE media (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  type        TEXT        NOT NULL,
  entity_id   UUID,
  bucket_key  TEXT        NOT NULL,
  public_url  TEXT        NOT NULL,
  sort_order  INT         NOT NULL DEFAULT 0,
  alt_text    TEXT,
  uploaded_by UUID        REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_bookings_barber_date  ON bookings(barber_id, date);
CREATE INDEX idx_bookings_user         ON bookings(user_id);
CREATE INDEX idx_bookings_status       ON bookings(status);
CREATE INDEX idx_media_type_entity     ON media(type, entity_id);
CREATE INDEX idx_media_sort            ON media(type, sort_order);
CREATE INDEX idx_refresh_tokens_user   ON refresh_tokens(user_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_refresh_tokens_user;
DROP INDEX IF EXISTS idx_media_sort;
DROP INDEX IF EXISTS idx_media_type_entity;
DROP INDEX IF EXISTS idx_bookings_status;
DROP INDEX IF EXISTS idx_bookings_user;
DROP INDEX IF EXISTS idx_bookings_barber_date;
DROP TABLE IF EXISTS media;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS booking_products;
DROP TABLE IF EXISTS bookings;
DROP TABLE IF EXISTS products;
DROP TABLE IF EXISTS services;
DROP TABLE IF EXISTS barber_schedules;
DROP TABLE IF EXISTS barbers;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd