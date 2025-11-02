CREATE TABLE manual_overrides (
    override_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    namespace TEXT NOT NULL,
    surface TEXT NOT NULL,
    action TEXT NOT NULL CHECK (action IN ('boost', 'suppress')),
    item_id TEXT NOT NULL,
    boost_value DOUBLE PRECISION,
    notes TEXT,
    created_by TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ,
    rule_id UUID REFERENCES rules(rule_id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'cancelled', 'expired')),
    cancelled_at TIMESTAMPTZ,
    cancelled_by TEXT
);

CREATE INDEX manual_overrides_org_idx ON manual_overrides (org_id, namespace, surface);
CREATE INDEX manual_overrides_status_idx ON manual_overrides (status, expires_at);
CREATE INDEX manual_overrides_rule_idx ON manual_overrides (rule_id);
