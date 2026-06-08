-- RE:FORM canonical schema
-- All IDs are UUIDs generated server-side.

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT        UNIQUE NOT NULL,
  password_hash TEXT        NOT NULL,
  first_name    TEXT,
  last_name     TEXT,
  phone         TEXT,
  role          TEXT        NOT NULL DEFAULT 'customer', -- customer | barber | founder
  reminder_opt  BOOLEAN     NOT NULL DEFAULT true,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- A barber is always a user first. One user → at most one barber profile.
CREATE TABLE barbers (
  id        UUID    PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id   UUID    UNIQUE NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  name      TEXT    NOT NULL,
  title     TEXT,
  bio       TEXT,
  num       TEXT,   -- display order label, e.g. '01'
  active    BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE barber_schedules (
  id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  barber_id  UUID NOT NULL REFERENCES barbers(id) ON DELETE CASCADE,
  dow        INT  NOT NULL, -- 0=Sun … 6=Sat
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
  reference    TEXT        UNIQUE NOT NULL, -- human-readable, e.g. RFB-A1B2C3
  user_id      UUID        REFERENCES users(id) ON DELETE SET NULL,
  barber_id    UUID        NOT NULL REFERENCES barbers(id),
  service_id   UUID        NOT NULL REFERENCES services(id),
  date         DATE        NOT NULL,
  time_start   TIME        NOT NULL,
  time_end     TIME        NOT NULL,
  status       TEXT        NOT NULL DEFAULT 'confirmed', -- confirmed | cancelled | completed
  created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
  cancelled_at TIMESTAMPTZ
);

CREATE TABLE booking_products (
  id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  booking_id  UUID NOT NULL REFERENCES bookings(id) ON DELETE CASCADE,
  product_id  UUID NOT NULL REFERENCES products(id),
  qty         INT  NOT NULL,
  price_pence INT  NOT NULL -- price snapshot at booking time
);

CREATE TABLE refresh_tokens (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id     UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash  TEXT        NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Single table covering all image asset types.
-- entity_id is NULL for carousel / gallery (not bound to a specific row).
CREATE TABLE media (
  id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  type        TEXT        NOT NULL, -- barber_photo | product_image | carousel | gallery
  entity_id   UUID,                 -- barbers.id or products.id when relevant
  bucket_key  TEXT        NOT NULL, -- R2 object key
  public_url  TEXT        NOT NULL,
  sort_order  INT         NOT NULL DEFAULT 0,
  alt_text    TEXT,
  uploaded_by UUID        REFERENCES users(id) ON DELETE SET NULL,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Indexes
CREATE INDEX idx_bookings_barber_date  ON bookings(barber_id, date);
CREATE INDEX idx_bookings_user         ON bookings(user_id);
CREATE INDEX idx_bookings_status       ON bookings(status);
CREATE INDEX idx_media_type_entity     ON media(type, entity_id);
CREATE INDEX idx_media_sort            ON media(type, sort_order);
CREATE INDEX idx_refresh_tokens_user   ON refresh_tokens(user_id);
