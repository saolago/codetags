package codetags

import "testing"

var table_NewInstance_Names = []string {
  "codetags", "CodeTags", "CODETAGS",
}

func Test_illegal_NewInstance_name(t *testing.T) {
  for _, name := range table_NewInstance_Names {
    _, err := NewInstance(name)
    if err == nil {
      t.Errorf("NewInstance(%s): must return a non-nil error", name)
    }
  }
}
