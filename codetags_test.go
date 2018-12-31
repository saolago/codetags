package codetags

import "os"
import "testing"
import "reflect"
import "github.com/stretchr/testify/assert"

func getFirstReturn(manager *TagManager, err error) *TagManager {
	return manager
}

func TestDefault(t *testing.T) {
	assert.Equal(t, Default(), getFirstReturn(GetInstance("CODETAGS")))
	assert.Equal(t, Default(), getFirstReturn(GetInstance("CodeTags")))
	assert.Equal(t, Default(), getFirstReturn(GetInstance("codetags")))
}

func TestNewInstance_invalid_name(t *testing.T) {
	name := ""
	_, err := NewInstance(name)
	if err == nil {
		t.Errorf("NewInstance('%s'): must return a non-nil error", name)
	}
}

func TestNewInstance_illegal_name(t *testing.T) {
	var tableNewInstanceNames = []string{
		"codetags", "CodeTags", "CODETAGS",
	}
	for _, name := range tableNewInstanceNames {
		_, err := NewInstance(name)
		if err == nil {
			t.Errorf("NewInstance('%s'): must return a non-nil error", name)
		}
	}
}

func TestInitialize(t *testing.T) {
	var tableInitializeCases = []struct {
		current  *Presets
		data     *Presets
		expected *Presets
	}{
		{
			current:  &Presets{"namespace": "ABC"},
			data:     &Presets{"namespace": "xyz", "version": "0.1.2"},
			expected: &Presets{"namespace": "XYZ", "version": "0.1.2"},
		},
	}
	for i, c := range tableInitializeCases {
		ct, _ := NewInstance("test", c.current)
		ctRef := ct.Initialize(c.data)
		if ctRef != ct {
			t.Errorf("testcase[%d] - output Ref is different with source Ref", i)
		}
		actual := ctRef.GetPresets()
		diffFields := []string{}
		for _, f := range []string{"namespace", "INCLUDED_TAGS", "EXCLUDED_TAGS", "version"} {
			if actual[f] != (*c.expected)[f] {
				diffFields = append(diffFields, f)
			}
		}
		if len(diffFields) > 0 {
			t.Errorf("testcase[%d] - different fields: %v", i, diffFields)
		}
	}
}

func TestRegister(t *testing.T) {
	var tableRegisterCases = []struct {
		presets      *Presets
		descriptors  []interface{}
		declaredTags []string
	}{
		{
			presets: &Presets{"version": "0.1.2"},
			descriptors: []interface{}{
				"tag-1",
				TagDescriptor{
					Name:    "tag-2",
					Enabled: false,
					Plan:    TagPlan{Enabled: false},
				},
				TagDescriptor{
					Name: "tag-3",
				},
			},
			declaredTags: []string{"tag-1", "tag-3"},
		},
	}
	for i, c := range tableRegisterCases {
		ct, _ := NewInstance("test", c.presets)
		ctRef := ct.Register(c.descriptors)
		if ctRef != ct {
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

	isacti, _ := NewInstance("isacti", &Presets{
		"namespace": "IsActive",
	})

	isacti.Register([]interface{}{"tag-1", "tag-2"})

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
	assert.True(t, isacti.IsActive([]interface{}{"abc", "xyz"}))
	assert.True(t, isacti.IsActive([]interface{}{"abc", "xyz"}, nil))
	assert.False(t, isacti.IsActive([]interface{}{"abc", "nil"}))
	assert.False(t, isacti.IsActive([]interface{}{"abc", "def", "nil"}))
	assert.False(t, isacti.IsActive([]interface{}{"abc", "def", "disabled"}))
	assert.False(t, isacti.IsActive([]interface{}{"abc", "123"}, []interface{}{"def", "456"}))
	// pre-defined tags are overridden by values of environment variables
	assert.True(t, isacti.IsActive("tag-1"))
	assert.True(t, isacti.IsActive("abc", "tag-1"))
	assert.True(t, isacti.IsActive("disabled", "tag-1"))
	assert.True(t, isacti.IsActive("tag-4"))
	assert.False(t, isacti.IsActive("tag-2"))
	assert.False(t, isacti.IsActive("tag-3"))
	assert.False(t, isacti.IsActive([]interface{}{nil, "tag-1"}))
	assert.False(t, isacti.IsActive([]interface{}{"nil", "tag-1"}))
	assert.False(t, isacti.IsActive("nil", "tag-3"))
	assert.False(t, isacti.IsActive("tag-3", "disabled"))
}
