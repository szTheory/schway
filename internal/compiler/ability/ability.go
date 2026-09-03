package ability

import (
	"fmt"

	"github.com/codename-lang/lang/internal/compiler/core"
)

var order = [...]core.Ability{
	core.AbilityCopy, core.AbilityDrop, core.AbilityShare, core.AbilitySend, core.AbilityEscape,
}

type Result struct {
	Granted           []core.Ability
	NegativeWitnesses []core.AbilityWitness
}

// Derive applies the sealed primitive and constructor rules. There is no
// production API for supplying or combining arbitrary ability masks.
func Derive(shape core.TypeRef) (Result, error) {
	result := Result{Granted: []core.Ability{}, NegativeWitnesses: []core.AbilityWitness{}}
	for _, candidate := range order {
		granted, path, err := derive(shape, candidate)
		if err != nil {
			return Result{}, err
		}
		if granted {
			result.Granted = append(result.Granted, candidate)
		} else {
			result.NegativeWitnesses = append(result.NegativeWitnesses, core.AbilityWitness{Ability: candidate, Path: path})
		}
	}
	return result, nil
}

func derive(shape core.TypeRef, candidate core.Ability) (bool, []string, error) {
	switch shape.Constructor {
	case "Byte":
		if len(shape.Arguments) != 0 {
			return false, nil, fmt.Errorf("Byte takes no type arguments")
		}
		return true, nil, nil
	case "Buffer":
		if len(shape.Arguments) != 0 {
			return false, nil, fmt.Errorf("Buffer takes no type arguments")
		}
		granted := candidate == core.AbilityDrop || candidate == core.AbilitySend || candidate == core.AbilityEscape
		if granted {
			return true, nil, nil
		}
		return false, []string{"Buffer"}, nil
	case "Box":
		if len(shape.Arguments) != 1 {
			return false, nil, fmt.Errorf("Box takes one type argument")
		}
		granted, path, err := derive(shape.Arguments[0], candidate)
		if !granted {
			path = append([]string{"Box.value"}, path...)
		}
		return granted, path, err
	case "Pair":
		if len(shape.Arguments) != 2 {
			return false, nil, fmt.Errorf("Pair takes two type arguments")
		}
		for index, argument := range shape.Arguments {
			granted, path, err := derive(argument, candidate)
			if err != nil {
				return false, nil, err
			}
			if !granted {
				field := "Pair.left"
				if index == 1 {
					field = "Pair.right"
				}
				return false, append([]string{field}, path...), nil
			}
		}
		return true, nil, nil
	default:
		return false, nil, fmt.Errorf("unknown type constructor %q", shape.Constructor)
	}
}

func Has(result Result, candidate core.Ability) bool {
	for _, granted := range result.Granted {
		if granted == candidate {
			return true
		}
	}
	return false
}
