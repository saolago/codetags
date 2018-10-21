// Package codetags is a simple feature toggle utility for Go.
// Developers could use this package to tag code blocks of a feature to 
// a declared label and turn on/off that feature by environment variables.
package codetags

import (
  "errors"
  "fmt"
  "os"
  "strings"
  "reflect"
  "regexp"
)
import "github.com/blang/semver"

const DEFAULT_NAMESPACE string = "CODETAGS"

type TagDescriptor struct {
  Name string
  Enabled interface{}
  Plan interface{}
  Note string
}

type TagPlan struct {
  Enabled interface{}
  MinBound interface{}
  MaxBound interface{}
}

type Presets = map[string]string

type TagManager struct {
  store struct {
    env map[string][]string
    declaredTags []string
    includedTags []string
    excludedTags []string
    cachedTags map[string]bool
  }
  presets Presets
}

var fieldOf_Initialize_opts_group1 = []string { "version" }

var fieldOf_Initialize_opts_group2 = []string {
  "namespace", "includedTagsLabel", "excludedTagsLabel",
}

func (c *TagManager) Initialize(opts *Presets) *TagManager {
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

// Registers the pre-defined tags
func (c *TagManager) Register(descriptors []interface{}) *TagManager {
  errs := []string {}
  defs := list_filter(descriptors, func(descriptor interface{}, idx int) bool {
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
    errs = append(errs, fmt.Sprintf(
      "descriptor#%d [%v] has invalid type (%s), must be a string or TagDescriptor type",
      idx, descriptor, reflect.TypeOf(descriptor).String(),
    ))
    return false
  })
  if len(errs) > 0 {
    panic(strings.Join(errs, "\n"))
  }
  tags := list_map(defs, func (info interface{}, index int) string {
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

func (c *TagManager) IsActive(tagexps ...interface{}) bool {
  return c.isArgumentsSatisfied(tagexps)
}

func (c *TagManager) isArgumentsSatisfied(tagexps []interface{}) bool {
  for _, tagexp := range tagexps {
    if c.evaluateExpression(tagexp) {
      return true
    }
  }
  return false
}

func (c *TagManager) isAllOfLabelsSatisfied(tagexp interface{}) bool {
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

func (c *TagManager) isAnyOfLabelsSatisfied(tagexp interface{}) bool {
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

func (c *TagManager) isNotOfLabelsSatisfied(tagexp interface{}) bool {
  return !c.evaluateExpression(tagexp)
}

func (c *TagManager) evaluateExpression(tagexp interface{}) bool {
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

func (c *TagManager) checkLabelActivated(label string) bool {
  if cachedVal, ok := c.store.cachedTags[label]; ok {
    return cachedVal
  }
  c.store.cachedTags[label] = c.forceCheckLabelActivated(label)
  return c.store.cachedTags[label]
}

func (c *TagManager) forceCheckLabelActivated(label string) bool {
  if list_contains(c.store.excludedTags, label) {
    return false
  }
  if list_contains(c.store.includedTags, label) {
    return true
  }
  return list_contains(c.store.declaredTags, label)
}

func (c *TagManager) GetDeclaredTags() []string {
  return list_clone(c.store.declaredTags)
}

func (c *TagManager) GetExcludedTags() []string {
  return list_clone(c.store.excludedTags)
}

func (c *TagManager) GetIncludedTags() []string {
  return list_clone(c.store.includedTags)
}

func (c *TagManager) GetPresets() Presets {
  return c.presets
}

func (c *TagManager) Reset() *TagManager {
  c.ClearCache()
  c.store.declaredTags = c.store.declaredTags[:0]
  for k := range c.presets {
    delete(c.presets, k)
  }
  return c
}

func (c *TagManager) ClearCache() *TagManager {
  for k := range c.store.cachedTags {
    delete(c.store.cachedTags, k)
  }
  return c.refreshEnv()
}

func (c *TagManager) refreshEnv() *TagManager {
  for k := range c.store.env {
    delete(c.store.env, k)
  }
  c.store.excludedTags = c.getEnv(c.getLabel("excludedTags"))
  c.store.includedTags = c.getEnv(c.getLabel("includedTags"))
  return c
}

func (c *TagManager) getEnv(label string) []string {
  if tags, ok := c.store.env[label]; ok {
    return tags
  }
  c.store.env[label] = stringToList(os.Getenv(label))
  return c.store.env[label]
}

func (c *TagManager) getLabel(tagType string) string {
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

var instances map[string]*TagManager = make(map[string]*TagManager)

var instance *TagManager = Default()

func Default() (*TagManager) {
  i, _ := GetInstance(DEFAULT_NAMESPACE)
  return i
}

func GetInstance(name string, opts ...*Presets) (*TagManager, error) {
  name = labelify(name)
  if instance, ok := instances[name]; ok {
    return instance, nil
  }
  if len(opts) > 0 {
    return createInstance(name, opts[0])
  } else {
    return createInstance(name, nil)
  }
}

func NewInstance(name string, opts ...*Presets) (*TagManager, error) {
  name = labelify(name)
  if name == DEFAULT_NAMESPACE {
    if _, ok := instances[name]; ok {
      return nil, errors.New(
        fmt.Sprintf("%s is default instance name. Please provides another name.",
          DEFAULT_NAMESPACE))
    }
  }
  if len(opts) > 0 {
    return createInstance(name, opts[0])
  } else {
    return createInstance(name, nil)
  }
}

func createInstance(name string, opts *Presets) (*TagManager, error) {
  if name == "" {
    return nil, errors.New(
      "The name of a codetags instance must be not empty")
  }
  c := &TagManager {}
  c.store.env = make(map[string][]string, 0)
  c.store.declaredTags = make([]string, 0)
  c.store.excludedTags = make([]string, 0)
  c.store.includedTags = make([]string, 0)
  c.store.cachedTags = make(map[string]bool, 0)
  c.presets = make(Presets)
  c.Initialize(opts)
  instances[name] = c
  return instances[name], nil
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

func list_filter(vs []interface{}, f func(interface{}, int) bool) []interface{} {
  vsf := make([]interface{}, 0)
  for i, v := range vs {
      if f(v, i) {
          vsf = append(vsf, v)
      }
  }
  return vsf
}

func list_map(vs []interface{}, f func(interface{}, int) string) []string {
  vsm := make([]string, len(vs))
  for i, v := range vs {
      vsm[i] = f(v, i)
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
