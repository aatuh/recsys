SELECT
    decision_id,
    org_id,
    ts,
    namespace,
    surface,
    request_id,
    user_hash,
    k,
    constraints,
    effective_config,
    bandit,
    candidates_pre,
    final_items,
    mmr_info,
    caps,
    extras
FROM rec_decisions
WHERE org_id = $1 AND decision_id = $2;
