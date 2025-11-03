CREATE TABLE recsys_item_factors (
    org_id uuid NOT NULL,
    namespace text NOT NULL,
    item_id text NOT NULL,
    factors vector(384) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, namespace, item_id)
);

CREATE INDEX recsys_item_factors_ns_idx
    ON recsys_item_factors (org_id, namespace);

CREATE TABLE recsys_user_factors (
    org_id uuid NOT NULL,
    namespace text NOT NULL,
    user_id text NOT NULL,
    factors vector(384) NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, namespace, user_id)
);

CREATE INDEX recsys_user_factors_ns_idx
    ON recsys_user_factors (org_id, namespace);
