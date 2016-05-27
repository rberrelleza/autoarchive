package util

import (
	"os"
	"strconv"
)

// Env represents the global environment.
var Env *GoEnv

// GoEnv is a type encapsulating the global environment.
type GoEnv struct {
	Mode  string
	Debug bool
}

// NewGoEnv creates a new GoEnv instance.
func NewGoEnv() *GoEnv {
	mode := os.Getenv("GO_ENV")
	if mode == "" {
		mode = "production"
		os.Setenv("GO_ENV", mode)
	}
	debug, _ := strconv.ParseBool(os.Getenv("GO_DEBUG"))
	return &GoEnv{
		Mode:  mode,
		Debug: debug,
	}
}

// IsProd tests if the current runtime is considered to be in `production` mode.
func (e *GoEnv) IsProd() bool {
	return e.Mode == "production"
}

// IsDev tests if the current runtime is considered to be in `development` mode.
func (e *GoEnv) IsDev() bool {
	return e.Mode == "development"
}

// GetString gets a string value from the environment.
func (e *GoEnv) GetString(key string) string {
	return os.Getenv(key)
}

// GetStringOr gets a string value from the environment, or a default value if it
// does not exist.
func (e *GoEnv) GetStringOr(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		val = def
	}
	return val
}

// GetInt gets an int value from the environment.
func (e *GoEnv) GetInt(key string) int {
	return e.GetIntOr(key, 0)
}

// GetIntOr gets an int value from the environment, or a default value if it
// does not exist.
func (e *GoEnv) GetIntOr(key string, def int) int {
	str := os.Getenv(key)
	val := def
	if str != "" {
		parsed, err := strconv.ParseInt(str, 10, 32)
		if err == nil {
			val = int(parsed)
		}
	}
	return val
}

func init() {
	Env = NewGoEnv()
}
