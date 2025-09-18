CREATE TYPE rule_action AS ENUM ('BLOCK', 'PIN', 'BOOST');
CREATE TYPE rule_target AS ENUM ('ITEM', 'TAG', 'BRAND', 'CATEGORY');

CREATE TABLE rules (
    rule_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    namespace TEXT NOT NULL,
    surface TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    action rule_action NOT NULL,
    target_type rule_target NOT NULL,
    target_key TEXT,
    item_ids TEXT[] DEFAULT '{}',
    boost_value DOUBLE PRECISION,
    max_pins INTEGER,
    segment_id TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    valid_from TIMESTAMPTZ,
    valid_until TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX rules_org_scope_idx ON rules (org_id, namespace, surface, enabled);
CREATE INDEX rules_validity_idx ON rules (valid_from, valid_until);
CREATE INDEX rules_segment_idx ON rules (org_id, namespace, segment_id);
