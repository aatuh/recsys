package types

import "github.com/aatuh/recsys-algo/rules"

// RuleAction enumerates supported rule actions.
type RuleAction = rules.RuleAction

const (
	RuleActionBlock = rules.RuleActionBlock
	RuleActionPin   = rules.RuleActionPin
	RuleActionBoost = rules.RuleActionBoost
)

// RuleTarget enumerates the supported rule target dimensions.
type RuleTarget = rules.RuleTarget

const (
	RuleTargetItem     = rules.RuleTargetItem
	RuleTargetTag      = rules.RuleTargetTag
	RuleTargetBrand    = rules.RuleTargetBrand
	RuleTargetCategory = rules.RuleTargetCategory
)

// Rule represents a deterministic merchandising rule.
type Rule = rules.Rule

// RuleListFilters captures optional filters for listing rules.
type RuleListFilters = rules.RuleListFilters

// RuleScope represents namespace/surface (+ optional segment) lookup key.
type RuleScope = rules.RuleScope
