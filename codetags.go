package codetags

import (
  "fmt"
  "strings"
  "regexp"
)

const DEFAULT_NAMESPACE string = "CODETAGS"
const conlog bool = true

type Presets = map[string]string

type codetags struct {
	store struct {
    env []string
    declaredTags []string
    includedTags []string
    excludedTags []string
    cachedTags map[string]bool
  }
  presets Presets
}

var fieldOf_Initialize_opts_group1 = []string{ "version" }

var fieldOf_Initialize_opts_group2 = []string {
  "namespace", "includedTagsLabel", "excludedTagsLabel",
}

func (c *codetags) Initialize(opts *Presets) *codetags {
  for _, key := range fieldOf_Initialize_opts_group1 {
    if val, ok := (*opts)[key]; ok {
      c.presets[key] = val
    }
  }
  for _, key := range fieldOf_Initialize_opts_group2 {
    if val, ok := (*opts)[key]; ok {
      c.presets[key] = labelify(val)
    }
  }
  return c
}

func (c *codetags) Register() *codetags {
  if conlog {
    fmt.Printf("Register() is invoked.\n")
  }
  return c
}

var instances map[string]codetags = make(map[string]codetags)

func NewInstance(name string, opts ...*Presets) (*codetags, error) {
  c := &codetags {}
  c.presets = make(Presets)
  if len(opts) > 0 {
    c.Initialize(opts[0])
  }
  return c, nil
}

var not_alphabet = regexp.MustCompile(`\W{1,}`)

func labelify(label string) string {
  return strings.ToUpper(not_alphabet.ReplaceAllString(strings.Trim(label, ` `), `_`))
}
