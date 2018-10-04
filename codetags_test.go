package codetags

import "fmt"
import "testing"
import "reflect"

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
    current Presets
    data Presets
    expected Presets
  }{
    {
      current: Presets { "namespace": "ABC" },
      data: Presets { "namespace": "xyz", "version": "0.1.2" },
      expected: Presets {"namespace": "XYZ", "version": "0.1.2"},
    },
  }
  for i, c := range table_Initialize_Cases {
    ct, _ := NewInstance("test", &c.current)
    ct_ref := ct.Initialize(&c.data)
    if ct_ref != ct {
      t.Errorf("testcase[%d] - output Ref is different with source Ref", i)
    }
    actual := ct_ref.presets
    diffFields := []string {}
    for _, f := range fieldOf_Initialize_opts_group1 {
      if actual[f] != c.expected[f] {
        diffFields = append(diffFields, f)
      }
    }
    for _, f := range fieldOf_Initialize_opts_group2 {
      if actual[f] != c.expected[f] {
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
        TagDescriptor{ Name: "tag-2", Enabled: false, Plan: TagPlan{ Enabled: false } },
        TagDescriptor{ Name: "tag-3" },
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
      t.Errorf("testcase[%d] - declaredTags[%v] is different with expected [%v]", i, actual, c.declaredTags)
    }
    if conlog {
      fmt.Printf("+> Input data: %v \n", c)
    }
  }
}