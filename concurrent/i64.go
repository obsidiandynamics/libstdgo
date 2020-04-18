package concurrent

// I64Condition is a predicate that checks whether the current (supplied) value meets some condition, returning
// true if the condition is met.
type I64Condition func(value int64) bool

// I64Not produces a logical inverse of the given condition.
func I64Not(cond I64Condition) I64Condition {
	return func(value int64) bool { return !cond(value) }
}

// I64Equal tests that the value equals a target value.
func I64Equal(target int64) I64Condition {
	return func(value int64) bool { return value == target }
}

// I64LessThan tests that the value is less than the given target value.
func I64LessThan(target int64) I64Condition {
	return func(value int64) bool { return value < target }
}

// I64LessThanOrEqual tests that the value is less than or equal to the given target value.
func I64LessThanOrEqual(target int64) I64Condition {
	return func(value int64) bool { return value <= target }
}

// I64GreaterThan tests that the value is greater than the given target value.
func I64GreaterThan(target int64) I64Condition {
	return func(value int64) bool { return value > target }
}

// I64GreaterThanOrEqual tests that the value is greater than or equal to the given target value.
func I64GreaterThanOrEqual(target int64) I64Condition {
	return func(value int64) bool { return value >= target }
}
