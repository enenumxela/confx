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

// Confx is a configuration registry.
type Confx struct {
	// Delimiter that separates a list of keys used to access a nested value in one go
	Delimeter string

	configFile string
	configType string

	overrides map[string]interface{}
	defaults  map[string]interface{}
	config    map[string]interface{}
}

var confx *Confx

func init() {
	confx = New()
}

// New initializes and retuns the initialized Confx instance.
func New() (confx *Confx) {
	confx = &Confx{}

	confx.Delimeter = "."

	confx.config = make(map[string]interface{})
	confx.defaults = make(map[string]interface{})
	confx.overrides = make(map[string]interface{})

	return
}

// SetOverride calls SetOverride method to sets the highest precedence comfiguration (confx.overrides).
func SetOverride(key string, value interface{}) { confx.SetOverride(key, value) }

// SetOverride method sets the highest precedence comfiguration (confx.overrides).
func (confx *Confx) SetOverride(key string, value interface{}) {
	key = strings.ToLower(key)
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, confx.Delimeter)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(confx.overrides, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// SetConfiguration calls SetConfiguration method to load the configuration file
func SetConfiguration(cf string) error { return confx.SetConfiguration(cf) }

// SetConfiguration method loads the configuration file
func (confx *Confx) SetConfiguration(cf string) (err error) {
	confx.configFile = cf

	ext := filepath.Ext(cf)
	if len(ext) < 1 {
		err = errors.New("unable to determine file type")
		return
	}

	confx.configType = ext[1:]

	if !stringInSlice(confx.configType, SupportedExts) {
		return
	}

	file, err := ioutil.ReadFile(cf)
	if err != nil {
		return
	}

	err = confx.unmarshalReader(bytes.NewReader(file), confx.config)
	if err != nil {
		return
	}

	return
}

// SetDefault calls SetDefault method to sets the lowest precedence comfiguration (confx.defaults).
func SetDefault(key string, value interface{}) { confx.SetDefault(key, value) }

// SetDefault method sets the lowest precedence comfiguration (confx.defaults).
func (confx *Confx) SetDefault(key string, value interface{}) {
	key = strings.ToLower(key)
	value = toCaseInsensitiveValue(value)

	path := strings.Split(key, confx.Delimeter)
	lastKey := strings.ToLower(path[len(path)-1])
	deepestMap := deepSearch(confx.defaults, path[0:len(path)-1])

	// set innermost value
	deepestMap[lastKey] = value
}

// searchMap recursively searches for a value for path in source map.
// Returns nil if not found.
// Note: This assumes that the path entries and map keys are lower cased.
func (confx *Confx) searchMap(source map[string]interface{}, path []string) interface{} {
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
			return confx.searchMap(to.StringMap(next), path[1:])
		case map[string]interface{}:
			// Type assertion is safe here since it is only reached
			// if the type of `next` is the same as the type being asserted
			return confx.searchMap(next.(map[string]interface{}), path[1:])
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
func (confx *Confx) searchIndexableWithPathPrefixes(source interface{}, path []string) interface{} {
	if len(path) == 0 {
		return source
	}

	// search for path prefixes, starting from the longest one
	for i := len(path); i > 0; i-- {
		prefixKey := strings.ToLower(strings.Join(path[0:i], confx.Delimeter))

		var val interface{}

		switch sourceIndexable := source.(type) {
		case []interface{}:
			val = confx.searchSliceWithPathPrefixes(sourceIndexable, prefixKey, i, path)
		case map[string]interface{}:
			val = confx.searchMapWithPathPrefixes(sourceIndexable, prefixKey, i, path)
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
func (confx *Confx) searchSliceWithPathPrefixes(
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
		return confx.searchIndexableWithPathPrefixes(to.StringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return confx.searchIndexableWithPathPrefixes(n, path[pathIndex:])
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
func (confx *Confx) searchMapWithPathPrefixes(
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
		return confx.searchIndexableWithPathPrefixes(to.StringMap(n), path[pathIndex:])
	case map[string]interface{}, []interface{}:
		return confx.searchIndexableWithPathPrefixes(n, path[pathIndex:])
	default:
		// got a value but nested key expected, do nothing and look for next prefix
	}

	// not found
	return nil
}

// GetConfigz gets the global Confx instance.
func GetConfigz() *Confx {
	return confx
}

func (confx *Confx) unmarshalReader(in io.Reader, c map[string]interface{}) error {
	buf := new(bytes.Buffer)
	buf.ReadFrom(in)

	switch strings.ToLower(confx.configType) {
	case "yaml", "yml":
		if err := yaml.Unmarshal(buf.Bytes(), &c); err != nil {
			return err
		}
	}

	insensitiviseMap(c)
	return nil
}
