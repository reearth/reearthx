package generator

import "sort"

type ResourceDefinition struct {
	Resource string
	Actions  []ActionDefinition
}

type ActionDefinition struct {
	Action    string
	Roles     []string
	Condition *Condition
}

type ResourceBuilder struct {
	serviceName string
	resources   map[string][]ActionDefinition
}

func NewResourceBuilder(serviceName string) *ResourceBuilder {
	return &ResourceBuilder{
		serviceName: serviceName,
		resources:   make(map[string][]ActionDefinition),
	}
}

func NewActionDefinition(action string, roles []string) ActionDefinition {
	return ActionDefinition{
		Action: action,
		Roles:  roles,
	}
}

func (b *ResourceBuilder) AddResource(resource string, actions []ActionDefinition) *ResourceBuilder {
	b.resources[resource] = actions
	return b
}

func (b *ResourceBuilder) Build() []ResourceDefinition {
	result := make([]ResourceDefinition, 0, len(b.resources))
	for resource, actions := range b.resources {
		result = append(result, ResourceDefinition{
			Resource: b.serviceName + ":" + resource,
			Actions:  actions,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Resource < result[j].Resource
	})
	return result
}

// NewActionDefinitionWithCondition creates an ActionDefinition with a condition
func NewActionDefinitionWithCondition(action string, roles []string, condition *Condition) ActionDefinition {
	return ActionDefinition{
		Action:    action,
		Roles:     roles,
		Condition: condition,
	}
}

// AllOf creates a condition where all expressions must be true (logical AND)
func AllOf(exprs ...string) *Condition {
	matches := make([]Match, len(exprs))
	for i, expr := range exprs {
		e := expr
		matches[i] = Match{Expr: &e}
	}
	return &Condition{
		Match: Match{
			All: &MatchExpressions{Of: matches},
		},
	}
}

// AnyOf creates a condition where at least one expression must be true (logical OR)
func AnyOf(exprs ...string) *Condition {
	matches := make([]Match, len(exprs))
	for i, expr := range exprs {
		e := expr
		matches[i] = Match{Expr: &e}
	}
	return &Condition{
		Match: Match{
			Any: &MatchExpressions{Of: matches},
		},
	}
}

// NoneOf creates a condition where none of the expressions should be true (logical negation)
func NoneOf(exprs ...string) *Condition {
	matches := make([]Match, len(exprs))
	for i, expr := range exprs {
		e := expr
		matches[i] = Match{Expr: &e}
	}
	return &Condition{
		Match: Match{
			None: &MatchExpressions{Of: matches},
		},
	}
}

// SimpleExpr creates a condition with a single expression
func SimpleExpr(expr string) *Condition {
	e := expr
	return &Condition{
		Match: Match{
			Expr: &e,
		},
	}
}
