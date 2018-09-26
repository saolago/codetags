package codetags

import "fmt"

const DEFAULT_NAMESPACE string = "CODETAGS"
const conlog bool = true

type Presets struct {
  namespace string
  includedTagsLabel string
  excludedTagsLabel string
  version string
}

type codetags struct {
	store struct {
    env []string
    declaredTags []string
    cachedTags map[string]bool
  }
  presets Presets
}

func (c *codetags) Initialize() *codetags {
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
  return &codetags {}, nil
}
