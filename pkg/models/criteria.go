package models

import (
	"encoding/json"
	"fmt"
)

// CriteriaObject represents the root wrapper for criteria
type CriteriaObject struct {
	Criteria json.RawMessage `json:"criteria"`
}

// QueryBuilder represents the main query builder structure
type QueryBuilder struct {
	ID         string      `json:"id"`
	Combinator string      `json:"combinator"` // "and" or "or"
	Not        bool        `json:"not"`
	Rules      []RuleGroup `json:"rules"`
}

// RuleGroup represents either a simple rule or a nested group of rules
type RuleGroup struct {
	ID          string      `json:"id"`
	Field       *string     `json:"field,omitempty"`       // Present for simple rules
	Operator    *string     `json:"operator,omitempty"`    // Present for simple rules
	Value       interface{} `json:"value,omitempty"`       // Present for simple rules
	ValueSource *string     `json:"valueSource,omitempty"` // Optional: "field" or "value"
	Combinator  *string     `json:"combinator,omitempty"`  // Present for grouped rules
	Not         *bool       `json:"not,omitempty"`         // Present for grouped rules
	Rules       []RuleGroup `json:"rules,omitempty"`       // Present for grouped rules
}

// IsSimpleRule returns true if this is a simple rule (not a group)
func (r *RuleGroup) IsSimpleRule() bool {
	return r.Field != nil && r.Operator != nil
}

// IsGroupRule returns true if this is a group rule (contains nested rules)
func (r *RuleGroup) IsGroupRule() bool {
	return r.Combinator != nil && r.Rules != nil
}

// SimpleRule represents a simple query rule
type SimpleRule struct {
	ID          string      `json:"id"`
	Field       string      `json:"field"`
	Operator    string      `json:"operator"`
	Value       interface{} `json:"value"`
	ValueSource *string     `json:"valueSource,omitempty"`
}

// GroupRule represents a group of rules with a combinator
type GroupRule struct {
	ID         string      `json:"id"`
	Combinator string      `json:"combinator"`
	Not        *bool       `json:"not,omitempty"`
	Rules      []RuleGroup `json:"rules"`
}

// ValidateQueryBuilder validates the query builder structure
func (qb *QueryBuilder) Validate() error {
	if qb.ID == "" {
		return fmt.Errorf("query builder ID is required")
	}

	if qb.Combinator != "and" && qb.Combinator != "or" {
		return fmt.Errorf("combinator must be 'and' or 'or'")
	}

	if len(qb.Rules) == 0 {
		return fmt.Errorf("at least one rule is required")
	}

	return qb.validateRules(qb.Rules)
}

// validateRules recursively validates all rules
func (qb *QueryBuilder) validateRules(rules []RuleGroup) error {
	for _, rule := range rules {
		if err := qb.validateRule(rule); err != nil {
			return err
		}
	}
	return nil
}

// validateRule validates a single rule
func (qb *QueryBuilder) validateRule(rule RuleGroup) error {
	if rule.ID == "" {
		return fmt.Errorf("rule ID is required")
	}

	if rule.IsSimpleRule() {
		// Validate simple rule
		if *rule.Field == "" {
			return fmt.Errorf("field is required for simple rule")
		}
		if *rule.Operator == "" {
			return fmt.Errorf("operator is required for simple rule")
		}
		if !IsOperatorSupported(*rule.Operator) {
			return fmt.Errorf("operator %s is not supported", *rule.Operator)
		}
		if rule.ValueSource != nil && !IsValueSourceSupported(*rule.ValueSource) {
			return fmt.Errorf("value source %s is not supported", *rule.ValueSource)
		}
		// Value can be empty for some operators like "isEmpty"
	} else if rule.IsGroupRule() {
		// Validate group rule
		if *rule.Combinator != "and" && *rule.Combinator != "or" {
			return fmt.Errorf("combinator must be 'and' or 'or' for group rule")
		}
		if len(rule.Rules) == 0 {
			return fmt.Errorf("group rule must contain at least one sub-rule")
		}
		// Recursively validate nested rules
		return qb.validateRules(rule.Rules)
	} else {
		return fmt.Errorf("rule must be either a simple rule or a group rule")
	}

	return nil
}

// SupportedOperators defines the supported query operators
var SupportedOperators = []string{
	"=", "!=", "<", "<=", ">", ">=",
	"contains", "beginsWith", "endsWith",
	"doesNotContain", "doesNotBeginWith", "doesNotEndWith",
	"null", "notNull",
	"in", "notIn", "between", "notBetween",
}

// SupportedValueSources defines the supported value sources
var SupportedValueSources = []string{
	"value", "field",
}

// IsOperatorSupported checks if an operator is supported
func IsOperatorSupported(operator string) bool {
	for _, op := range SupportedOperators {
		if op == operator {
			return true
		}
	}
	return false
}

// IsValueSourceSupported checks if a value source is supported
func IsValueSourceSupported(valueSource string) bool {
	for _, vs := range SupportedValueSources {
		if vs == valueSource {
			return true
		}
	}
	return false
}
