package saga

import "testing"

func TestDefinitionValidate(t *testing.T) {
	def := Definition{Steps: []Step{{Name: "step-1"}, {Name: "step-2"}}}
	if err := def.Validate(); err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
}

func TestDefinitionValidateDuplicate(t *testing.T) {
	def := Definition{Steps: []Step{{Name: "step"}, {Name: "step"}}}
	if err := def.Validate(); err == nil {
		t.Fatalf("expected error")
	}
}

func TestDefinitionValidateMissing(t *testing.T) {
	def := Definition{}
	if err := def.Validate(); err == nil {
		t.Fatalf("expected error")
	}
}
