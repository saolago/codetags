package codetags

import "fmt"
import "os"
import "testing"
import "reflect"
import "github.com/stretchr/testify/assert"

func Test_illegal_NewInstance_name(t *testing.T) {
  t.Skip()
  var table_NewInstance_Names = []string {
    "codetags", "CodeTags", "CODETAGS",
  }
  for _, name := range table_NewInstance_Names {
    _, err := NewInstance(name)
    if err == nil {
      t.Errorf("NewInstance('%s'): must return a non-nil error", name)
    }
  }
}

func Test_labelify(t *testing.T) {
  var table_labelify_cases = []struct {
    label string
    expected string
  }{
    { label: "", expected: "" },
    { label: "Hello  world", expected: "HELLO_WORLD" },
    { label: "Underscore_with 123", expected: "UNDERSCORE_WITH_123" },
    { label: "user@example.com", expected: "USER_EXAMPLE_COM" },
  }
  for _, c := range table_labelify_cases {
    actual := labelify(c.label)
    if actual != c.expected {
      t.Errorf("labelify(%s): expected %s, actual %s", c.label, c.expected, actual)
    }
  }
}

func Test_Initialize(t *testing.T) {
  var table_Initialize_Cases = []struct {
    current *Presets
    data *Presets
    expected *Presets
  }{
    {
      current: &Presets { "namespace": "ABC" },
      data: &Presets { "namespace": "xyz", "version": "0.1.2" },
      expected: &Presets {"namespace": "XYZ", "version": "0.1.2"},
    },
  }
  for i, c := range table_Initialize_Cases {
    ct, _ := NewInstance("test", c.current)
    ct_ref := ct.Initialize(c.data)
    if ct_ref != ct {
      t.Errorf("testcase[%d] - output Ref is different with source Ref", i)
    }
    actual := ct_ref.presets
    diffFields := []string {}
    for _, f := range fieldOf_Initialize_opts_group1 {
      if actual[f] != (*c.expected)[f] {
        diffFields = append(diffFields, f)
      }
    }
    for _, f := range fieldOf_Initialize_opts_group2 {
      if actual[f] != (*c.expected)[f] {
        diffFields = append(diffFields, f)
      }
    }
    if len(diffFields) > 0 {
      t.Errorf("testcase[%d] - different fields: %v", i, diffFields)
    }
    if conlog {
      fmt.Printf("+> Output: %v \n", diffFields)
    }
  }
}

func Test_Register(t *testing.T) {
  var table_Register_Cases = []struct {
    presets *Presets
    descriptors []interface{}
    declaredTags []string
  }{
    {
      presets: &Presets{ "version": "0.1.2" },
      descriptors: []interface{} {
        "tag-1",
        TagDescriptor{
          Name: "tag-2",
          Enabled: false,
          Plan: TagPlan{ Enabled: false },
        },
        TagDescriptor{
          Name: "tag-3",
        },
      },
      declaredTags: []string { "tag-1", "tag-3" },
    },
  }
  for i, c := range table_Register_Cases {
    ct, _ := NewInstance("test", c.presets)
    ct_ref := ct.Register(c.descriptors)
    if ct_ref != ct {
      t.Errorf("testcase[%d] - output Ref is different with source Ref", i)
    }
    actual := ct.GetDeclaredTags()
    if !reflect.DeepEqual(actual, c.declaredTags) {
      t.Errorf("testcase[%d] - actual[%v] is different with expected [%v]", i, actual, c.declaredTags)
    }
    if conlog {
      fmt.Printf("+> Input data: %v \n", c)
    }
  }
}

func Test_evaluateExpression(t *testing.T) {
  os.Setenv("ISACTIVE_INCLUDED_TAGS", "abc, def, xyz, tag-4")
  os.Setenv("ISACTIVE_EXCLUDED_TAGS", "disabled, tag-2")

  isacti, _ := NewInstance("isacti", &Presets{
    "namespace": "IsActive",
  })

  isacti.Register([]interface{} {"tag-1", "tag-2"})

  assert.True(t, isacti.evaluateExpression("abc"))
  assert.True(t, isacti.evaluateExpression("xyz"))
  assert.True(t, isacti.evaluateExpression([]string {"abc", "xyz"}))
  assert.True(t, isacti.evaluateExpression([]interface{} {"abc", "xyz"}))
  assert.True(t, isacti.evaluateExpression(map[string]interface{} {
    "$all": []interface{} { "abc", "xyz" },
    "$not": "not-found",
    "$any": []interface{} { "tag-0", "tag-4" },
  }))
  assert.True(t, isacti.evaluateExpression(map[string]interface{} {
    "$all": []string { "abc", "xyz" },
    "$not": "not-found",
    "$any": []string { "tag-0", "tag-4" },
  }))
  assert.True(t, isacti.evaluateExpression(map[string]interface{} {
    "$not": "tag-0",
    "$all": []interface{} {
      "tag-1", "tag-4",
    },
  }))
  assert.True(t, isacti.evaluateExpression(map[string]interface{} {
    "$all": []interface{} { "abc", "xyz", map[string]interface{} {
      "$not": "tag-0",
      "$all": []interface{} {
        "tag-1", "tag-4",
      },
    }},
    "$not": "not-found",
  }))
  assert.False(t, isacti.evaluateExpression(nil))
  assert.False(t, isacti.evaluateExpression("nil"))
}

func Test_IsActive(t *testing.T) {
  os.Setenv("ISACTIVE_INCLUDED_TAGS", "abc, def, xyz, tag-4")
  os.Setenv("ISACTIVE_EXCLUDED_TAGS", "disabled, tag-2")

  isacti, _ := NewInstance("isacti", &Presets{
    "namespace": "IsActive",
  })

  isacti.Register([]interface{} {"tag-1", "tag-2"})

  includedExpected := []string{"abc", "def", "xyz", "tag-4"}
  if !reflect.DeepEqual(isacti.GetIncludedTags(), includedExpected) {
    t.Errorf("includedTags [%v] is different with expected [%v]", isacti.GetIncludedTags(), includedExpected)
  }

  excludedExpected := []string{"disabled", "tag-2"}
  if !reflect.DeepEqual(isacti.GetExcludedTags(), excludedExpected) {
    t.Errorf("excludedTags [%v] is different with expected [%v]", isacti.GetExcludedTags(), excludedExpected)
  }
  // An arguments-list presents the OR conditional operator
  assert.True(t, isacti.IsActive("abc"))
  assert.True(t, isacti.IsActive("xyz"))
  assert.True(t, isacti.IsActive("abc", "disabled"))
  assert.True(t, isacti.IsActive("disabled", "abc"))
  assert.True(t, isacti.IsActive("abc", "nil"))
  assert.True(t, isacti.IsActive("undefined", "abc", "nil"))
  assert.False(t, isacti.IsActive())
  assert.False(t, isacti.IsActive(nil))
  assert.False(t, isacti.IsActive("disabled"))
  assert.False(t, isacti.IsActive("nil"))
  assert.False(t, isacti.IsActive("disabled", "nil"))
  // An array argument presents the AND conditional operator
  assert.True(t, isacti.IsActive([]interface{} { "abc", "xyz" }))
  assert.True(t, isacti.IsActive([]interface{} { "abc", "xyz" }, nil))
  assert.False(t, isacti.IsActive([]interface{} { "abc", "nil" }))
  assert.False(t, isacti.IsActive([]interface{} { "abc", "def", "nil" }))
  assert.False(t, isacti.IsActive([]interface{} { "abc", "def", "disabled" }))
  assert.False(t, isacti.IsActive([]interface{} { "abc", "123" }, []interface{} { "def", "456" }))
  // pre-defined tags are overridden by values of environment variables
  assert.True(t, isacti.IsActive("tag-1"))
  assert.True(t, isacti.IsActive("abc", "tag-1"))
  assert.True(t, isacti.IsActive("disabled", "tag-1"))
  assert.True(t, isacti.IsActive("tag-4"))
  assert.False(t, isacti.IsActive("tag-2"))
  assert.False(t, isacti.IsActive("tag-3"))
  assert.False(t, isacti.IsActive([]interface{} { nil, "tag-1" }))
  assert.False(t, isacti.IsActive([]interface{} { "nil", "tag-1" }))
  assert.False(t, isacti.IsActive("nil", "tag-3"))
  assert.False(t, isacti.IsActive("tag-3", "disabled"))
}
