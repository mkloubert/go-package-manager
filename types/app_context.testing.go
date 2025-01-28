package types

import "flag"

// app.IsRunningInTests() - `true` if tests are running
func (app *AppContext) IsRunningInTests() bool {
	return flag.Lookup("test.v") != nil
}
