package compare

// Eqer can be used to determine if this value is equal to the other.
type Eqer interface {
	Eq(other interface{}) bool
}

// ProbablyEqer is an equal check that may return false positives, but never
// a false negative.
type ProbablyEqer interface {
	ProbablyEq(other interface{}) bool
}

// Comparer can be used to compare two values.
// This will be used when using the le, ge etc. operators in the templates.
// Compare returns -1 if the given version is less than, 0 if equal and 1 if greater than
// the running version.
type Comparer interface {
	Compare(other interface{}) int
}
