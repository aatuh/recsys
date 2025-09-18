-- Segment profiles and deterministic rule-based segments

CREATE TABLE public.segment_profiles (
    org_id uuid NOT NULL,
    namespace text NOT NULL,
    profile_id text NOT NULL,
    description text NOT NULL DEFAULT '',
    blend_alpha double precision NOT NULL,
    blend_beta double precision NOT NULL,
    blend_gamma double precision NOT NULL,
    mmr_lambda double precision NOT NULL,
    brand_cap integer NOT NULL,
    category_cap integer NOT NULL,
    profile_boost double precision NOT NULL,
    profile_window_days double precision NOT NULL,
    profile_top_n integer NOT NULL,
    half_life_days double precision NOT NULL,
    co_vis_window_days integer NOT NULL,
    purchased_window_days integer NOT NULL,
    rule_exclude_events boolean NOT NULL,
    exclude_event_types smallint[] NOT NULL DEFAULT '{}'::smallint[],
    brand_tag_prefixes text[] NOT NULL DEFAULT '{}'::text[],
    category_tag_prefixes text[] NOT NULL DEFAULT '{}'::text[],
    popularity_fanout integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, namespace, profile_id)
);

CREATE TABLE public.segments (
    org_id uuid NOT NULL,
    namespace text NOT NULL,
    segment_id text NOT NULL,
    name text NOT NULL,
    priority integer NOT NULL,
    active boolean NOT NULL DEFAULT true,
    profile_id text NOT NULL,
    description text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, namespace, segment_id),
    CONSTRAINT segments_profile_fk FOREIGN KEY (org_id, namespace, profile_id)
        REFERENCES public.segment_profiles (org_id, namespace, profile_id)
        ON UPDATE CASCADE ON DELETE RESTRICT
);

CREATE UNIQUE INDEX segments_priority_uq
    ON public.segments (org_id, namespace, priority, segment_id);

CREATE TABLE public.segment_rules (
    org_id uuid NOT NULL,
    namespace text NOT NULL,
    segment_id text NOT NULL,
    rule_id bigserial NOT NULL,
    rule jsonb NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    description text NOT NULL DEFAULT '',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (org_id, namespace, segment_id, rule_id),
    CONSTRAINT segment_rules_segment_fk FOREIGN KEY (org_id, namespace, segment_id)
        REFERENCES public.segments (org_id, namespace, segment_id)
        ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE INDEX segment_rules_enabled_idx
    ON public.segment_rules (org_id, namespace, segment_id)
    WHERE enabled = true;
