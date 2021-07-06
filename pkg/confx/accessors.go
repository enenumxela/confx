package confx

import (
	"strings"

	"github.com/enenumxela/to/pkg/to"
)

func (confx *Confx) find(key string) (value interface{}) {
	path := strings.Split(key, confx.Delimeter)

	// Overrides
	value = confx.searchMap(confx.overrides, path)
	if value != nil {
		return
	}

	// configuration file
	value = confx.searchIndexableWithPathPrefixes(confx.config, path)
	if value != nil {
		return
	}

	// defaults
	value = confx.searchMap(confx.defaults, path)
	if value != nil {
		return
	}

	return
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} { return confx.Get(key) }

func (confx *Confx) Get(key string) interface{} {
	lcaseKey := strings.ToLower(key)
	val := confx.find(lcaseKey)
	if val == nil {
		return nil
	}

	return val
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return confx.getString(key) }

func (confx *Confx) getString(key string) string {
	return to.String(confx.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return confx.getInt(key) }

func (confx *Confx) getInt(key string) int {
	return to.Int(confx.Get(key))
}
