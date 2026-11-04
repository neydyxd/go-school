package l02interfacesinternals

type ValidationError struct {
	Field  string
	Reason string
}

func (e *ValidationError) Error() string { return e.Field + ": " + e.Reason }

func checkAge(age int) *ValidationError {
	if age < 0 || age > 150 {
		return &ValidationError{Field: "age", Reason: "вне диапазона 0-150"}
	}
	return nil
}

func ValidateAge(age int) error {
	if ve := checkAge(age); ve != nil {
		return ve
	}
	return nil
}
