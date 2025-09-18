SELECT
    decision_id,
    org_id,
    ts,
    namespace,
    surface,
    request_id,
    user_hash,
    k,
    final_items,
    extras
FROM rec_decisions
WHERE org_id = $1
  AND namespace = $2
  AND ($3::timestamptz IS NULL OR ts >= $3)
  AND ($4::timestamptz IS NULL OR ts <= $4)
  AND ($5::text IS NULL OR user_hash = $5)
  AND ($6::text IS NULL OR request_id = $6)
ORDER BY ts DESC
LIMIT $7;
