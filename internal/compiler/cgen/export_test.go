package cgen

// Test-only accessors. These make the emitters' reserved sets and the two
// identifier-namespace constructors visible to the external cgen_test package
// without exporting them from the production API.

var (
	MatchFixedNames  = matchFixedNames
	LinearFixedNames = linearFixedNames
)

func CName(name string) string  { return cName(name) }
func CLocal(name string) string { return cLocal(name) }

func AllocateFrom(reserved []string, preferred, category string, ordinal int) string {
	return newCNames(reserved...).allocate(preferred, category, ordinal)
}
