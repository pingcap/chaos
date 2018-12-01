package core

// Checker checks a history of operations.
type Checker interface {
	// Check a series of operations with the given model.
	// Return false or error if operations do not satisfy the model.
	Check(m Model, ops []Operation) (bool, error)

	// Name returns the unique name for the checker.
	Name() string
}

// NoopChecker is a noop checker.
type NoopChecker struct{}

// Check impls Checker.
func (NoopChecker) Check(m Model, ops []Operation) (bool, error) {
	return true, nil
}

// Name impls Checker.
func (NoopChecker) Name() string {
	return "NoopChecker"
}
