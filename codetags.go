package codetags

import (
  "fmt"
  "strings"
  "reflect"
  "regexp"
)
import "github.com/blang/semver"

const DEFAULT_NAMESPACE string = "CODETAGS"
const conlog bool = true

type TagDescriptor struct {
  Name string
  Enabled interface{}
  Plan interface{}
}

type TagPlan struct {
  Enabled interface{}
  MinBound interface{}
  MaxBound interface{}
}

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

var name_TagDescriptor string = typeof(TagDescriptor{})
var name_TagPlan string = typeof(TagPlan{})

func (c *codetags) Register(descriptors []interface{}) *codetags {
  defs := list_filter(descriptors, func(descriptor interface{}) bool {
    descriptorType := typeof(descriptor)
    if descriptorType == "string" {
      return true
    }
    if descriptorType == name_TagDescriptor {
      info := descriptor.(TagDescriptor)
      if info.Plan != nil && typeof(info.Plan) == name_TagPlan {
        plan := info.Plan.(TagPlan)
        if plan.Enabled != nil && typeof(plan.Enabled) == "bool" {
          if versionStr, ok := c.presets["version"]; ok {
            validated := true
            satisfied := true
            version, versionErr := semver.Make(versionStr)
            validated = validated && (versionErr == nil)
            if plan.MinBound != nil && typeof(plan.MinBound) == "string" {
              minBound, minBoundErr := semver.Make(plan.MinBound.(string))
              validated = validated && (minBoundErr == nil)
              satisfied = satisfied && (version.Compare(minBound) >= 0)
            }
            if plan.MaxBound != nil && typeof(plan.MaxBound) == "string" {
              maxBound, maxBoundErr := semver.Make(plan.MaxBound.(string))
              validated = validated && (maxBoundErr == nil)
              satisfied = satisfied && (version.Compare(maxBound) < 0)
            }
            if validated {
              if satisfied {
                return plan.Enabled.(bool)
              } else {
                if info.Enabled != nil && typeof(info.Enabled) == "bool" {
                  return info.Enabled.(bool)
                }
                return !plan.Enabled.(bool)
              }
            }
          }
        }
      }
      if info.Enabled != nil && typeof(info.Enabled) == "bool" {
        return info.Enabled.(bool)
      }
      return true
    }
    return false
  })
  tags := list_map(defs, func (info interface{}) string {
    if typeof(info) == name_TagDescriptor {
      descriptor := info.(TagDescriptor)
      return descriptor.Name
    }
    return info.(string)
  })
  for _, tag := range tags {
    if !list_contains(c.store.declaredTags, tag) {
      c.store.declaredTags = append(c.store.declaredTags, tag)
    }
  }
  if conlog {
    fmt.Printf("Register() is invoked: %v.\n", c.store.declaredTags)
  }
  return c
}

func (c *codetags) GetDeclaredTags() []string {
  return c.store.declaredTags
}

var instances map[string]codetags = make(map[string]codetags)

func NewInstance(name string, opts ...*Presets) (*codetags, error) {
  c := &codetags {}
  c.store.declaredTags = make([]string, 0)
  c.store.excludedTags = make([]string, 0)
  c.store.includedTags = make([]string, 0)
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

func typeof(v interface{}) string {
  return reflect.TypeOf(v).String()
}

func list_index(vs []string, t string) int {
  for i, v := range vs {
      if v == t {
          return i
      }
  }
  return -1
}

func list_contains(vs []string, t string) bool {
  return list_index(vs, t) >= 0
}

func list_filter(vs []interface{}, f func(interface{}) bool) []interface{} {
  vsf := make([]interface{}, 0)
  for _, v := range vs {
      if f(v) {
          vsf = append(vsf, v)
      }
  }
  return vsf
}

func list_map(vs []interface{}, f func(interface{}) string) []string {
  vsm := make([]string, len(vs))
  for i, v := range vs {
      vsm[i] = f(v)
  }
  return vsm
}
