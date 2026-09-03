package origins

import (
	"fmt"
	"sort"
)

type Registry struct {
	types map[string]TypeSpec
}

func NewRegistry(specs []TypeSpec) (Registry, error) {
	registry := Registry{types: make(map[string]TypeSpec, len(specs))}
	for _, spec := range specs {
		if spec.Name == "" {
			return Registry{}, fmt.Errorf("type name is empty")
		}
		if _, exists := registry.types[spec.Name]; exists {
			return Registry{}, fmt.Errorf("duplicate type %q", spec.Name)
		}
		for ability, indexes := range spec.Conditional {
			if !knownAbility(ability) {
				return Registry{}, fmt.Errorf("type %s has unknown ability %q", spec.Name, ability)
			}
			for _, index := range indexes {
				if index < 0 || index >= spec.Parameters {
					return Registry{}, fmt.Errorf("type %s ability %s references parameter %d", spec.Name, ability, index)
				}
			}
		}
		registry.types[spec.Name] = spec
	}
	return registry, nil
}

func (r Registry) Has(expr TypeExpr, ability Ability, substitutions map[string]TypeExpr) (bool, error) {
	if replacement, ok := substitutions[expr.Name]; ok && len(expr.Args) == 0 {
		return r.Has(replacement, ability, substitutions)
	}
	spec, ok := r.types[expr.Name]
	if !ok {
		return false, fmt.Errorf("unknown type %q", expr.Name)
	}
	if len(expr.Args) != spec.Parameters {
		return false, fmt.Errorf("type %s has %d arguments, want %d", expr.Name, len(expr.Args), spec.Parameters)
	}
	if containsAbility(spec.Base, ability) {
		return true, nil
	}
	indexes, conditional := spec.Conditional[ability]
	if !conditional {
		return false, nil
	}
	for _, index := range indexes {
		has, err := r.Has(expr.Args[index], ability, substitutions)
		if err != nil || !has {
			return false, err
		}
	}
	return true, nil
}

func (r Registry) Abilities(expr TypeExpr, substitutions map[string]TypeExpr) ([]Ability, error) {
	var abilities []Ability
	for _, ability := range AbilityOrder {
		has, err := r.Has(expr, ability, substitutions)
		if err != nil {
			return nil, err
		}
		if has {
			abilities = append(abilities, ability)
		}
	}
	return abilities, nil
}

func Substitute(expr TypeExpr, substitutions map[string]TypeExpr) TypeExpr {
	if replacement, ok := substitutions[expr.Name]; ok && len(expr.Args) == 0 {
		return replacement
	}
	result := TypeExpr{Name: expr.Name}
	for _, arg := range expr.Args {
		result.Args = append(result.Args, Substitute(arg, substitutions))
	}
	return result
}

func canonicalAbilities(values []Ability) []Ability {
	seen := map[Ability]bool{}
	var result []Ability
	for _, ability := range AbilityOrder {
		if containsAbility(values, ability) && !seen[ability] {
			seen[ability] = true
			result = append(result, ability)
		}
	}
	return result
}

func canonicalStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	write := 0
	for _, value := range result {
		if write == 0 || result[write-1] != value {
			result[write] = value
			write++
		}
	}
	return result[:write]
}

func containsAbility(values []Ability, target Ability) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func knownAbility(ability Ability) bool {
	return containsAbility(AbilityOrder, ability)
}
