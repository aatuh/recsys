# Candidate generation vs ranking (lean)

Candidates: high-recall sources (popularity, cooc, embeddings)
Ranking: scoring + constraints + diversification to produce the final top-K

Candidate overrides:
- `candidates.include_ids` acts as an allow-list (only these items may appear).
- `candidates.exclude_ids` always removes items from the final list.

Rules:
- Pin rules can inject items even if they are not in the base candidate pool.
