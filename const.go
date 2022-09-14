package terror

// Const defines a type of error without capturing the line where it was defined. To use it in a
// function, call terror.Wrap func to capture the line where it is being wrapped, within the
// function. It can be returned directly to improve performance in tight loops if necessary.
// Compared to combining errors.New, it can be assigned to a const.
type Const string

// Error implements the conventional interface for representing an error condition.
func (err Const) Error() string {
	return string(err)
}
