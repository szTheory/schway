// Package surface exposes stable, user-facing descriptions of constructs that
// the current checked language deliberately refuses.
package surface

// Refusal identifies one supported diagnostic boundary for a currently
// unsupported source form. Hints remain prose; they do not grant repair
// authority.
type Refusal struct {
	Code string
	Form string
	Hint string
}

var refusals = [...]Refusal{
	{
		Code: "surface.unsupported_logical_and",
		Form: "&&",
		Hint: "Logical && is unsupported; use explicit if/else control flow.",
	},
	{
		Code: "surface.unsupported_logical_or",
		Form: "||",
		Hint: "Logical || is unsupported; use explicit if/else control flow.",
	},
	{
		Code: "surface.unsupported_logical_not",
		Form: "!",
		Hint: "Logical ! is unsupported; use explicit if/else control flow.",
	},
	{
		Code: "surface.unsupported_shift_amount",
		Form: "non-literal, negative, or out-of-range shift amount",
		Hint: "Shift amounts must be integer literals from 0 to 63; write a literal in that range.",
	},
	{
		Code: "surface.unsupported_parameter_arity",
		Form: "more than one declaration parameter",
		Hint: "Declarations require exactly one parameter; pass one admitted value.",
	},
	{
		Code: "surface.unsupported_if_without_else",
		Form: "if statement without else block",
		Hint: "An if statement requires an else block.",
	},
	{
		Code: "surface.unsupported_bool_literal_tail",
		Form: "unbound true or false in a result tail",
		Hint: "Bool literals are unsupported in result tails; return an initialized value allowed by the declared result type.",
	},
	{
		Code: "surface.unsupported_subtraction",
		Form: "-",
		Hint: "Subtraction is unsupported; no unchecked emulation is recommended.",
	},
}

// Refusals returns the current catalog in stable source order.
func Refusals() []Refusal {
	return append([]Refusal(nil), refusals[:]...)
}

// LookupRefusal finds one catalog entry by its stable diagnostic code.
func LookupRefusal(code string) (Refusal, bool) {
	for _, refusal := range refusals {
		if refusal.Code == code {
			return refusal, true
		}
	}
	return Refusal{}, false
}
