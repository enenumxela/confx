// Each source takes precedence as follows
//
// 1. overrides
// 2. configuration file
// 3. defaults

package confx

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/enenumxela/to/pkg/to"
	"gopkg.in/yaml.v3"
)

// SupportedExts is a slice of supported configuration file extensions
var SupportedExts = []string{"yaml", "yml"}

// Configz is a configuration reistry.
type Configz struct {
	Delimeter string

	configFile string
	configType string

	overrides map[string]interface{}
	defaults  map[string]interface{}
	config    map[string]interface{}
}

var configz *Configz

func init() {
	configz = New()
}

// New initializes and retuns the initialized Configz instance.
func New() (configz *Configz) {
	configz = &Configz{}

	configz.Delimeter = "."

	configz.overrides = make(map[string]interface{})
	configz.defaults = make(map[string]interface{})
	configz.config = make(map[string]interface{})
	return
}

// SetOverride calls SetOverride method to sets the highest precedence comfiguration (configz.overrides).
func SetOverride(key string, value interface{}) { configz.SetOverride(key, value) }

// SetOverride method sets the highest precedence comfiguration (configz.overrides).
func (configz *Configz) SetOverride(key string, value interface{}) {
	key = strings.ToLower(key)
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, configz.Delimeter)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(configz.overrides, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// SetConfiguration calls SetConfiguration method to load the configuration file
func SetConfiguration(cf string) error { return configz.SetConfiguration(cf) }

// SetConfiguration method loads the configuration file
func (configz *Configz) SetConfiguration(cf string) (err error) {
	configz.configFile = cf

	ext := filepath.Ext(cf)
	if len(ext) < 1 {
		err = errors.New("unable to determine file type")
		return
	}

	configz.configType = ext[1:]

	if !stringInSlice(configz.configType, SupportedExts) {
		return
	}

	file, err := ioutil.ReadFile(cf)
	if err != nil {
		return
	}

	err = configz.unmarshalReader(bytes.NewReader(file), configz.config)
	if err != nil {
		return
	}

	return
}

// SetDefault calls SetDefault method to sets the lowest precedence comfiguration (configz.defaults).
func SetDefault(key string, value interface{}) { configz.SetDefault(key, value) }

// SetDefault method sets the lowest precedence comfiguration (configz.defaults).
func (configz *Configz) SetDefault(key string, value interface{}) {
	key = strings.ToLower(key)
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, configz.Delimeter)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(configz.defaults, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

func (configz *Configz) find(key string) (value interface{}) {
	path := strings.Split(key, configz.Delimeter)

	// Overrides
	value = configz.searchMap(configz.overrides, path)
	if value != nil {
		return
	}

	// configuration file
	value = configz.searchIndexableWithPathPrefixes(configz.config, path)
	if value != nil {
		return
	}

	// defaults
	value = configz.searchMap(configz.defaults, path)
	if value != nil {
		return
	}

	return
}

// searchMap recursively searches for a value for path in source map.
// Returns nil if not found.
// Note: This assumes that the path entries and map keys are lower cased.
func (configz *Configz) searchMap(source map[string]interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	next, ok := source[path[0]]
	if ok {
		// Fast path
		if len(path) == 1 {
			return next
		}

		// Nested case
		switch next.(type) {
		case map[interface{}]interface{}:
			return configz.searchMap(to.StringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return configz.searchMap(next.(map[string]interface{}), path[1:])
		default:
			// got a value but nested key expected, return "nil" for not found
			return nil
		}
	}
	return nil
}

// searchIndexableWithPathPrefixes recursively searches for a value for path in source map/slice.
//
// While searchMap() considers each path element as a single map key or slice index, this
// function searches for, and prioritizes, merged path elements.
// e.g., if in the source, "foo" is defined with a sub-key "bar", and "foo.bar"
// is also defined, this latter value is returned for path ["foo", "bar"].
//
// This should be useful only at config level (other maps may not contain dots
// in their keys).
//
// Note: This assumes that the path entries and map keys are lower cased.
func (configz *Configz) searchIndexableWithPathPrefixes(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], configz.Delimeter))

		var val interface{}

		switch sourceIndexable := source.(type) {
		case []interface{}:
			val = configz.searchSliceWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		case map[string]interface{}:
			val = configz.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		}

		if val != nil {
			return val
		}
	}

	// not found
	return nil
}

// searchSliceWithPathPrefixes searches for a value for path in sourceSlice
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (configz *Configz) searchSliceWithPathPrefixes(
	sourceSlice []interface{},
	prefixKey string,
	pathIndex int,
	path []string,
) interface{} {
	// if the prefixKey is not a number or it is out of bounds of the slice
	index, err := strconv.Atoi(prefixKey)
	if err != nil || len(sourceSlice) <= index {
		return nil
	}

	next := sourceSlice[index]

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	switch n := next.(type) {
	case map[interface{}]interface{}:
		return configz.searchIndexableWithPathPrefixes(to.StringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return configz.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// searchMapWithPathPrefixes searches for a value for path in sourceMap
//
// This function is part of the searchIndexableWithPathPrefixes recurring search and
// should not be called directly from functions other than searchIndexableWithPathPrefixes.
func (configz *Configz) searchMapWithPathPrefixes(
	sourceMap map[string]interface{},
	prefixKey string,
	pathIndex int,
	path []string,
) interface{} {
	next, ok := sourceMap[prefixKey]
	if !ok {
		return nil
	}

	// Fast path
	if pathIndex == len(path) {
		return next
	}

	// Nested case
	switch n := next.(type) {
	case map[interface{}]interface{}:
		return configz.searchIndexableWithPathPrefixes(to.StringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return configz.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// GetConfigz gets the global Configz instance.
func GetConfigz() *Configz {
	return configz
}

func (configz *Configz) unmarshalReader(in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	switch strings.ToLower(configz.configType) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(buf.Bytes(), &c); err != nil {
			return err
		}
	}

	insensitiviseMap(c)
	return nil
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} { return configz.Get(key) }

func (configz *Configz) Get(key string) interface{} {
	lcaseKey := strings.ToLower(key)
	val := configz.find(lcaseKey)
	if val == nil {
		return nil
	}

	return val
}

// GetString returns the value associated with the key as a string.
func GetString(key string) string { return configz.getString(key) }

func (configz *Configz) getString(key string) string {
	return to.String(configz.Get(key))
}

// GetInt returns the value associated with the key as an integer.
func GetInt(key string) int { return configz.getInt(key) }

func (configz *Configz) getInt(key string) int {
	return to.Int(configz.Get(key))
}
