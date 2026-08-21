CREATE TABLE facilities (id text PRIMARY KEY, payload jsonb NOT NULL, version bigint NOT NULL, updated_at timestamptz NOT NULL);
CREATE TABLE programs (id text PRIMARY KEY, facility_id text NOT NULL REFERENCES facilities(id), payload jsonb NOT NULL, version bigint NOT NULL, updated_at timestamptz NOT NULL);
CREATE TABLE work_windows (id text PRIMARY KEY, facility_id text NOT NULL REFERENCES facilities(id), idempotency_key text NOT NULL UNIQUE, payload jsonb NOT NULL, version bigint NOT NULL, updated_at timestamptz NOT NULL);
CREATE TABLE executions (id text PRIMARY KEY, window_id text NOT NULL REFERENCES work_windows(id), idempotency_key text NOT NULL UNIQUE, payload jsonb NOT NULL, version bigint NOT NULL, updated_at timestamptz NOT NULL);
CREATE TABLE incidents (id text PRIMARY KEY, facility_id text NOT NULL REFERENCES facilities(id), payload jsonb NOT NULL, version bigint NOT NULL, updated_at timestamptz NOT NULL);
CREATE TABLE audit_events (id text PRIMARY KEY, aggregate text NOT NULL, aggregate_id text NOT NULL, action text NOT NULL, actor_id text NOT NULL, payload jsonb NOT NULL, occurred_at timestamptz NOT NULL);
CREATE INDEX audit_events_aggregate_idx ON audit_events(aggregate, aggregate_id, occurred_at);
