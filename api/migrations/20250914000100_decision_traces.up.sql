CREATE TABLE IF NOT EXISTS rec_decisions (
    decision_id UUID PRIMARY KEY,
    org_id UUID NOT NULL,
    ts TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    namespace TEXT NOT NULL,
    surface TEXT,
    request_id TEXT,
    user_hash TEXT,
    k INT,
    constraints JSONB,
    effective_config JSONB NOT NULL,
    bandit JSONB,
    candidates_pre JSONB NOT NULL,
    final_items JSONB NOT NULL,
    mmr_info JSONB,
    caps JSONB,
    extras JSONB
);

CREATE INDEX IF NOT EXISTS idx_recdec_ns_ts ON rec_decisions (namespace, ts);
CREATE INDEX IF NOT EXISTS idx_recdec_org_ns_ts ON rec_decisions (org_id, namespace, ts);
CREATE INDEX IF NOT EXISTS idx_recdec_req ON rec_decisions (request_id);
CREATE INDEX IF NOT EXISTS idx_recdec_user_ts ON rec_decisions (user_hash, ts);
