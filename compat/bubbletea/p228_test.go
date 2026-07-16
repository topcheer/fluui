package bubbletea

import "testing"

// P228: Verify value-type models work with NewProgram + Run + type assertion
// ggcode: model := newResumePickerModel(...); finalModel, _ := NewProgram(model).Run(); result, _ := finalModel.(resumePickerModel)

type valueModel struct {
	step int
}

func (m valueModel) Init() Cmd                  { return Quit() }
func (m valueModel) Update(msg Msg) (Model, Cmd) { return m, nil }
func (m valueModel) View() View                  { return NewView("") }

func TestNewProgramWithValueModel_Run_TypeAssert_P228(t *testing.T) {
	m := valueModel{step: 42}
	p := NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	// Type assertion to value type (not pointer)
	result, ok := finalModel.(valueModel)
	if !ok {
		t.Fatalf("type assertion failed, got %T", finalModel)
	}
	if result.step != 42 {
		t.Errorf("expected step=42, got %d", result.step)
	}
}

// Also test pointer model with type assertion (onboard pattern)
type ptrModel struct {
	val string
}

func (m *ptrModel) Init() Cmd                  { return Quit() }
func (m *ptrModel) Update(msg Msg) (Model, Cmd) { return m, nil }
func (m *ptrModel) View() View                  { return NewView("") }

func TestNewProgramWithPointerModel_Run_TypeAssert_P228(t *testing.T) {
	m := &ptrModel{val: "hello"}
	p := NewProgram(m)
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	result, ok := finalModel.(*ptrModel)
	if !ok {
		t.Fatalf("type assertion failed, got %T", finalModel)
	}
	if result.val != "hello" {
		t.Errorf("expected val=hello, got %q", result.val)
	}
}

// Test one-liner pattern: NewProgram(model).Run()
func TestNewProgramChainedRun_P228(t *testing.T) {
	m := valueModel{step: 99}
	finalModel, err := NewProgram(m).Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	if finalModel == nil {
		t.Error("should return non-nil model")
	}
}

// Test ProgramOption chaining
func TestNewProgramWithOptions_P228(t *testing.T) {
	m := valueModel{}
	p := NewProgram(m, WithoutSignals(), WithoutRenderer())
	finalModel, err := p.Run()
	if err != nil {
		t.Errorf("Run error: %v", err)
	}
	if finalModel == nil {
		t.Error("should return non-nil model")
	}
}
