package codetags_test

import "fmt"
import "os"
import "testing"
import "reflect"
import "github.com/stretchr/testify/assert"
import "github.com/saolago/codetags"

func getFirstReturn(manager *codetags.TagManager, err error) *codetags.TagManager {
  return manager
}

func ExampleDefault() {
  tagHandler := codetags.Default()

  // Reload environment variables
  os.Setenv("CODETAGS_INCLUDED_TAGS", "tag-1,tag-2")
  os.Setenv("CODETAGS_EXCLUDED_TAGS", "tag-2,tag-3")
  tagHandler.Reset()

  fmt.Printf("includedTags: %v\n", tagHandler.GetIncludedTags())
  fmt.Printf("excludedTags: %v\n", tagHandler.GetExcludedTags())
  // Output:
  // includedTags: [tag-1 tag-2]
  // excludedTags: [tag-2 tag-3]
}

func TestDefault(t *testing.T) {
  assert.Equal(t, codetags.Default(), getFirstReturn(codetags.GetInstance("CODETAGS")))
  assert.Equal(t, codetags.Default(), getFirstReturn(codetags.GetInstance("CodeTags")))
  assert.Equal(t, codetags.Default(), getFirstReturn(codetags.GetInstance("codetags")))
}

func ExampleNewInstance_01() {
  os.Setenv("MISSION_INCLUDED_TAGS", "tag-1,tag-2")
  os.Setenv("MISSION_EXCLUDED_TAGS", "tag-2,tag-3")

  tagHandler, err := codetags.NewInstance("myname", &codetags.Presets{
    "namespace": "Mission",
  })

  fmt.Printf("Error: %v\n", err)
  fmt.Printf("includedTags: %v\n", tagHandler.GetIncludedTags())
  fmt.Printf("excludedTags: %v\n", tagHandler.GetExcludedTags())
  // Output:
  // Error: <nil>
  // includedTags: [tag-1 tag-2]
  // excludedTags: [tag-2 tag-3]
}

func ExampleNewInstance_02() {
  tagHandler, err := codetags.NewInstance("CODETAGS", &codetags.Presets{
    "namespace": "Mission",
  })
  fmt.Printf("Handler: %v, Error: %v\n", tagHandler, err)
  // Output:
  // Handler: <nil>, Error: CODETAGS is default instance name. Please provides another name.
}

func TestNewInstance_invalid_name(t *testing.T) {
  name := ""
  _, err := codetags.NewInstance(name)
  if err == nil {
    t.Errorf("NewInstance('%s'): must return a non-nil error", name)
  }
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
  }
}

func ExampleTagManager_Register_02() {
  defer func() {
    if r := recover(); r != nil {
      fmt.Println(r)
    }
  }()

  tagHandler := codetags.Default().Reset()
  tagHandler.Register([]interface{} {
    "feature-1",
    codetags.TagDescriptor {
      Name: "feature-2",
    },
    1024,
    codetags.TagDescriptor {
      Name: "feature-4",
      Enabled: true,
    },
    true,
  })

  // Output:
  // descriptor#2 [1024] has invalid type (int), must be a string or TagDescriptor type
  // descriptor#4 [true] has invalid type (bool), must be a string or TagDescriptor type
}

// A simple usage: defined a list of tags that can be turned on/off defaultly by Enabled value.
func ExampleTagManager_Register_01() {
  tagHandler := codetags.Default().Reset()

  tagHandler.Register([]interface{} {
    "feature-1",
    codetags.TagDescriptor {
      Name: "feature-2",
    },
    codetags.TagDescriptor {
      Name: "feature-3",
      Enabled: false,
    },
    codetags.TagDescriptor {
      Name: "feature-4",
      Enabled: true,
    },
    codetags.TagDescriptor {
      Name: "feature-5",
      Plan: codetags.TagPlan {
        Enabled: true,
      },
    },
  })

  fmt.Printf("declaredTags: %v", tagHandler.GetDeclaredTags())
  // Output:
  // declaredTags: [feature-1 feature-2 feature-4 feature-5]
}

func ExampleTagManager_Register_03() {
  tagHandler := codetags.Default().Reset()

  tagHandler.Initialize(&codetags.Presets {
    "version": "0.1.7",
  })

  tagHandler.Register([]interface{} {
    codetags.TagDescriptor {
      Name: "feature-11",
      Plan: codetags.TagPlan {
        Enabled: true,
      },
    },
    codetags.TagDescriptor {
      Name: "feature-12",
      Plan: codetags.TagPlan {
        Enabled: true,
        MinBound: "0.1.2",
      },
    },
    codetags.TagDescriptor {
      Name: "feature-13",
      Plan: codetags.TagPlan {
        Enabled: true,
        MinBound: "0.1.2",
        MaxBound: "0.1.6",
      },
    },
    codetags.TagDescriptor {
      Name: "feature-14",
      Plan: codetags.TagPlan {
        Enabled: false,
        MinBound: "0.1.2",
        MaxBound: "0.1.6",
      },
    },
    codetags.TagDescriptor {
      Name: "feature-15",
      Plan: codetags.TagPlan {
        Enabled: true,
        MinBound: "0.1.8",
      },
    },
    codetags.TagDescriptor {
      Name: "feature-16",
      Plan: codetags.TagPlan {
        Enabled: false,
        MinBound: "0.1.9",
      },
    },
  })

  fmt.Printf("declaredTags: %v", tagHandler.GetDeclaredTags())
  // Output:
  // declaredTags: [feature-11 feature-12 feature-14 feature-16]
}

func ExampleTagManager_Register_04() {
  defer func() {
    if r := recover(); r != nil {
      fmt.Println(r)
    }
  }()

  tagHandler := codetags.Default().Reset()

  tagHandler.Initialize(&codetags.Presets {
    "version": "0.1.7",
  })

  tagHandler.Register([]interface{} {
    "feature-11",
    codetags.TagDescriptor {
      Name: "feature-11",
      Plan: codetags.TagPlan {
        Enabled: true,
      },
    },
    codetags.TagDescriptor {
      Name: "feature-12",
      Plan: codetags.TagPlan {
        Enabled: true,
        MinBound: "0.1.2",
      },
    },
    codetags.TagDescriptor {
      Name: "feature-13",
      Plan: codetags.TagPlan {
        Enabled: true,
        MinBound: "0.1.2",
        MaxBound: "0.1.6",
      },
    },
    "feature-13",
    codetags.TagDescriptor {
      Name: "feature-14",
      Plan: codetags.TagPlan {
        Enabled: false,
        MinBound: "0.1.2",
        MaxBound: "0.1.6",
      },
    },
    "feature-14",
  })

  // Output:
  // Tag [feature-11] is declared more than one time
  // Tag [feature-14] is declared more than one time
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
