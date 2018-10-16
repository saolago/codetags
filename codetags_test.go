package codetags_test

import "fmt"
import "os"
import "testing"
import "reflect"
import "github.com/stretchr/testify/assert"
import "github.com/saolago/codetags"

const conlog bool = true

func ExampleDefault() {
  os.Setenv("CODETAGS_INCLUDED_TAGS", "tag-1,tag-2")
  os.Setenv("CODETAGS_EXCLUDED_TAGS", "tag-2,tag-3")

  checker := codetags.Default()
  checker.Reset()
  
  fmt.Printf("includedTags: %v\n", checker.GetIncludedTags())
  fmt.Printf("excludedTags: %v\n", checker.GetExcludedTags())
  // Output:
  // includedTags: [tag-1 tag-2]
  // excludedTags: [tag-2 tag-3]
}

func TestNewInstance_illegal_name(t *testing.T) {
  var table_NewInstance_Names = []string {
    "codetags", "CodeTags", "CODETAGS",
  }
  for _, name := range table_NewInstance_Names {
    _, err := codetags.NewInstance(name)
    if err == nil {
      t.Errorf("NewInstance('%s'): must return a non-nil error", name)
    }
  }
}

func TestInitialize(t *testing.T) {
  var table_Initialize_Cases = []struct {
    current *codetags.Presets
    data *codetags.Presets
    expected *codetags.Presets
  }{
    {
      current: &codetags.Presets { "namespace": "ABC" },
      data: &codetags.Presets { "namespace": "xyz", "version": "0.1.2" },
      expected: &codetags.Presets {"namespace": "XYZ", "version": "0.1.2"},
    },
  }
  for i, c := range table_Initialize_Cases {
    ct, _ := codetags.NewInstance("test", c.current)
    ct_ref := ct.Initialize(c.data)
    if ct_ref != ct {
      t.Errorf("testcase[%d] - output Ref is different with source Ref", i)
    }
    actual := ct_ref.GetPresets()
    diffFields := []string {}
    for _, f := range []string { "namespace", "includedTagsLabel", "excludedTagsLabel", "version" } {
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

func TestRegister(t *testing.T) {
  var table_Register_Cases = []struct {
    presets *codetags.Presets
    descriptors []interface{}
    declaredTags []string
  }{
    {
      presets: &codetags.Presets{ "version": "0.1.2" },
      descriptors: []interface{} {
        "tag-1",
        codetags.TagDescriptor{
          Name: "tag-2",
          Enabled: false,
          Plan: codetags.TagPlan{ Enabled: false },
        },
        codetags.TagDescriptor{
          Name: "tag-3",
        },
      },
      declaredTags: []string { "tag-1", "tag-3" },
    },
  }
  for i, c := range table_Register_Cases {
    ct, _ := codetags.NewInstance("test", c.presets)
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

func TestIsActive(t *testing.T) {
  os.Setenv("ISACTIVE_INCLUDED_TAGS", "abc, def, xyz, tag-4")
  os.Setenv("ISACTIVE_EXCLUDED_TAGS", "disabled, tag-2")

  isacti, _ := codetags.NewInstance("isacti", &codetags.Presets{
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
