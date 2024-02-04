# Glossary

A small shared vocabulary used throughout the suite docs.

## Artifact

An immutable, version-addressed blob produced offline (for example: popularity, co-visitation, embeddings) and consumed
by `recsys-service`.

## Manifest

A small document that maps artifact types to artifact URIs for a `(tenant, surface)` pair. In artifact mode, the “current
manifest pointer” is what you ship and roll back.

## Candidate

An item considered for ranking. Candidate generation prioritizes recall: “what items should we even consider?”

## Ranking

Scoring and ordering candidates to produce the final top-K list. Ranking prioritizes precision: “which of these are
best?”

## Exposure

An event that records what items were shown (and in what order) for a single recommendation request.

## Outcome

An event that records what the user did after an exposure (click, conversion, etc.). Outcomes are attributed to exposures
by `request_id`.

## Assignment

An event that records experiment bucket membership (experiment id + variant) used for A/B analysis.

## Request ID

A correlation identifier that ties together “recommend” responses, exposure logs, and outcome events. In this suite it is
the primary join key for evaluation.

## Segment

A cohort/slice label (default: `default`) used for rule scoping and evaluation breakdowns (guest vs returning, locale,
etc.).

## Control plane

The admin APIs and versioned documents (tenant config + rules) that control serving behavior without redeploying code.

## Ship

Promote a candidate change to “current” in a safe, reversible way (for example: update the manifest pointer, or update
tenant config/rules versions).

## Rollback

Revert “current” to a last-known-good version (config/rules rollback or manifest pointer rollback).

## Namespace

An application-defined bucket of recommendation logic/data. In this suite, `surface` typically acts as the namespace for
signals and rules.

## Surface

Where recommendations are shown (home, PDP, cart, …). Surface names should be stable; they scope signals, rules, and
evaluation slices.

## Tenant

An organization boundary for configuration and data isolation.
