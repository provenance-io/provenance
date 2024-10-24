package testutil

import "os"

const TestnetEnvVar = "PIO_TESTNET"

// UnsetTestnetEnvVar will unset the PIO_TESTNET environment variable and return a deferrable that will put it back.
//
// Go runs tests inside an environment that might already have some environment variables defined. E.g. if you
// have `export PIO_TESTNET=true` in your environment, and run `make test`, then, when a test runs, it will
// start with a `PIO_TESTNET` value of `true`. But that can mess up some tests that expect to start without a
// PIO_TESTNET env var set.
//
// For individual test cases, you should use t.Setenv for changing environment variables.
// This exists because t.Setenv can't be used to unset an environment variable.
//
// Standard usage: defer testutil.UnsetTestnetEnvVar()()
func UnsetTestnetEnvVar() func() {
	if origVal, ok := os.LookupEnv(TestnetEnvVar); ok {
		os.Unsetenv(TestnetEnvVar)
		return func() {
			os.Setenv(TestnetEnvVar, origVal)
		}
	}
	return func() {
		os.Unsetenv(TestnetEnvVar)
	}
}
