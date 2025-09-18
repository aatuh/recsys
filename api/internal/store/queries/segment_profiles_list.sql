-- name: segment_profiles_list
SELECT
    profile_id,
    description,
    blend_alpha,
    blend_beta,
    blend_gamma,
    mmr_lambda,
    brand_cap,
    category_cap,
    profile_boost,
    profile_window_days,
    profile_top_n,
    half_life_days,
    co_vis_window_days,
    purchased_window_days,
    rule_exclude_events,
    exclude_event_types,
    brand_tag_prefixes,
    category_tag_prefixes,
    popularity_fanout,
    created_at,
    updated_at
FROM public.segment_profiles
WHERE org_id = $1 AND namespace = $2
ORDER BY profile_id;
