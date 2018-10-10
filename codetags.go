package codetags

import (
  "os"
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
    env map[string][]string
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
  if opts != nil {
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
  }
  return c.refreshEnv()
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
  return c
}

func (c *codetags) IsActive(tagexps ...interface{}) bool {
  return c.isArgumentsSatisfied(tagexps)
}

func (c *codetags) isArgumentsSatisfied(tagexps []interface{}) bool {
  for _, tagexp := range tagexps {
    if c.evaluateExpression(tagexp) {
      return true
    }
  }
  return false
}

func (c *codetags) isAllOfLabelsSatisfied(tagexp interface{}) bool {
  expType := reflect.TypeOf(tagexp)
  if expType.Kind().String() == "slice" {
    expElemKind := expType.Elem().Kind().String()
    if expElemKind == "string" {
      subexps := tagexp.([]string)
      for _, subexp := range subexps {
        if !c.checkLabelActivated(subexp) {
          return false
        }
      }
      return true
    }
    if expElemKind == "interface" {
      subexps := tagexp.([]interface{})
      for _, subexp := range subexps {
        if !c.evaluateExpression(subexp) {
          return false
        }
      }
      return true
    }
    return false
  }
  return c.evaluateExpression(tagexp)
}

func (c *codetags) isAnyOfLabelsSatisfied(tagexp interface{}) bool {
  expType := reflect.TypeOf(tagexp)
  if expType.Kind().String() == "slice" {
    expElemKind := expType.Elem().Kind().String()
    if expElemKind == "string" {
      subexps := tagexp.([]string)
      for _, subexp := range subexps {
        if c.checkLabelActivated(subexp) {
          return true
        }
      }
      return false
    }
    if expElemKind == "interface" {
      subexps := tagexp.([]interface{})
      for _, subexp := range subexps {
        if c.evaluateExpression(subexp) {
          return true
        }
      }
      return false
    }
    return false
  }
  return c.evaluateExpression(tagexp)
}

func (c *codetags) isNotOfLabelsSatisfied(tagexp interface{}) bool {
  return !c.evaluateExpression(tagexp)
}

func (c *codetags) evaluateExpression(tagexp interface{}) bool {
  if tagexp == nil {
    return false
  }
  expType := reflect.TypeOf(tagexp)
  expTypeKind := expType.Kind().String()
  // type: string
  if expTypeKind == "string" {
    return c.checkLabelActivated(tagexp.(string))
  }
  // type: array of anythings
  if expTypeKind == "slice" {
    return c.isAllOfLabelsSatisfied(tagexp)
  }
  // type: map of anythings
  if expTypeKind == "map" {
    expElem := expType.Elem()
    expElemKind := expElem.Kind().String()
    if (expElemKind == "interface") {
      subexps := tagexp.(map[string]interface{})
      for op, subexp := range subexps {
        switch (op) {
        case "$not":
          if !c.isNotOfLabelsSatisfied(subexp) {
            return false
          }
        case "$all":
          if !c.isAllOfLabelsSatisfied(subexp) {
            return false
          }
        case "$any":
          if !c.isAnyOfLabelsSatisfied(subexp) {
            return false
          }
        default:
          return false
        }
      }
      return true
    }
  }
  // type: unknown
  return false
}

func (c *codetags) checkLabelActivated(label string) bool {
  if cachedVal, ok := c.store.cachedTags[label]; ok {
    return cachedVal
  }
  c.store.cachedTags[label] = c.forceCheckLabelActivated(label)
  return c.store.cachedTags[label]
}

func (c *codetags) forceCheckLabelActivated(label string) bool {
  if list_contains(c.store.excludedTags, label) {
    return false
  }
  if list_contains(c.store.includedTags, label) {
    return true
  }
  return list_contains(c.store.declaredTags, label)
}

func (c *codetags) GetDeclaredTags() []string {
  return list_clone(c.store.declaredTags)
}

func (c *codetags) GetExcludedTags() []string {
  return list_clone(c.store.excludedTags)
}

func (c *codetags) GetIncludedTags() []string {
  return list_clone(c.store.includedTags)
}

func (c *codetags) Reset() *codetags {
  c.ClearCache()
  c.store.declaredTags = c.store.declaredTags[:0]
  for k := range c.presets {
    delete(c.presets, k)
  }
  return c
}

func (c *codetags) ClearCache() *codetags {
  for k := range c.store.cachedTags {
    delete(c.store.cachedTags, k)
  }
  return c.refreshEnv()
}

func (c *codetags) refreshEnv() *codetags {
  for k := range c.store.env {
    delete(c.store.env, k)
  }
  c.store.excludedTags = c.getEnv(c.getLabel("excludedTags"))
  c.store.includedTags = c.getEnv(c.getLabel("includedTags"))
  return c
}

func (c *codetags) getEnv(label string) []string {
  if tags, ok := c.store.env[label]; ok {
    return tags
  }
  c.store.env[label] = stringToList(os.Getenv(label))
  return c.store.env[label]
}

func (c *codetags) getLabel(tagType string) string {
  label := ""
  if namespace, ok := c.presets["namespace"]; ok && len(namespace) > 0 {
    label = namespace + "_"
  } else {
    label = DEFAULT_NAMESPACE + "_"
  }
  switch (tagType) {
  case "excludedTags":
    if tagLabel, ok := c.presets["excludedTagsLabel"]; ok && len(tagLabel) > 0 {
      label = label + tagLabel
    } else {
      label = label + "EXCLUDED_TAGS"
    }
  case "includedTags":
    if tagLabel, ok := c.presets["includedTagsLabel"]; ok && len(tagLabel) > 0 {
      label = label + tagLabel
    } else {
      label = label + "INCLUDED_TAGS"
    }
  default:
    if tagLabel, ok := c.presets[tagType]; ok && len(tagLabel) > 0 {
      label = label + tagLabel
    } else {
      label = label + labelify(tagType)
    }
  }
  return label
}

var instances map[string]codetags = make(map[string]codetags)

func NewInstance(name string, opts ...*Presets) (*codetags, error) {
  c := &codetags {}
  c.store.env = make(map[string][]string, 0)
  c.store.declaredTags = make([]string, 0)
  c.store.excludedTags = make([]string, 0)
  c.store.includedTags = make([]string, 0)
  c.store.cachedTags = make(map[string]bool, 0)
  c.presets = make(Presets)
  if len(opts) > 0 {
    c.Initialize(opts[0])
  } else {
    c.Initialize(nil)
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

func list_clone(ss []string) []string {
  if ss == nil {
    return make([]string, 0)
  }
  ts := make([]string, len(ss))
  copy(ts, ss)
  return ts
}

func stringToList(label string) []string {
  tags := make([]string, 0)
  strs := strings.Split(label, ",")
  for _, str := range strs {
    s := strings.Trim(str, ` `)
    if len(s) > 0 {
      tags = append(tags, s)
    }
  }
  return tags
}
