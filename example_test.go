package codetags_test

import "fmt"
import "os"
import "github.com/saolago/codetags"

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
