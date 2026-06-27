package block

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/topcheer/fluui/component"
	"github.com/topcheer/fluui/internal/buffer"
)

func TestNewWorkflowBlock(t *testing.T) {
	wb := NewWorkflowBlock("Deploy Agent")
	if wb.Title() != "Deploy Agent" {
		t.Errorf("Title = %q, want 'Deploy Agent'", wb.Title())
	}
	if wb.Type() != TypeWorkflow {
		t.Errorf("Type = %v, want TypeWorkflow", wb.Type())
	}
	if wb.StepCount() != 0 {
		t.Errorf("StepCount = %d, want 0", wb.StepCount())
	}
	if wb.Progress() != 0 {
		t.Errorf("Progress = %f, want 0", wb.Progress())
	}
}

func TestNewWorkflowBlock_IDGenerated(t *testing.T) {
	wb := NewWorkflowBlock("test")
	if wb.ID() == "" {
		t.Error("ID should be non-empty")
	}
	if !strings.HasPrefix(wb.ID(), "wf-") {
		t.Errorf("ID = %q, want prefix 'wf-'", wb.ID())
	}
}

func TestWorkflowAddStep(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("step1", "First step")
	wb.AddStep("step2", "Second step")

	if wb.StepCount() != 2 {
		t.Fatalf("StepCount = %d, want 2", wb.StepCount())
	}
	steps := wb.Steps()
	if steps[0].Name != "step1" || steps[0].Description != "First step" {
		t.Errorf("step[0] = %+v", steps[0])
	}
	if steps[0].Status != StepPending {
		t.Errorf("step[0] Status = %v, want pending", steps[0].Status)
	}
	if steps[1].Name != "step2" {
		t.Errorf("step[1].Name = %q", steps[1].Name)
	}
}

func TestWorkflowSetStepStatus(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	wb.SetStepStatus(0, StepRunning)
	steps := wb.Steps()
	if steps[0].Status != StepRunning {
		t.Errorf("step[0] Status = %v, want running", steps[0].Status)
	}

	wb.SetStepStatus(0, StepDone)
	steps = wb.Steps()
	if steps[0].Status != StepDone {
		t.Errorf("step[0] Status = %v, want done", steps[0].Status)
	}
	if steps[0].Duration <= 0 {
		t.Error("Duration should be > 0 after done")
	}
}

func TestWorkflowSetStepStatus_OutOfRange(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	// Negative index — should not panic
	wb.SetStepStatus(-1, StepDone)
	// Out of range — should not panic
	wb.SetStepStatus(99, StepDone)

	if wb.Steps()[0].Status != StepPending {
		t.Error("step should still be pending")
	}
}

func TestWorkflowProgress(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")
	wb.AddStep("s3", "")
	wb.AddStep("s4", "")

	// 0 completed
	if got := wb.Progress(); got != 0 {
		t.Errorf("Progress = %f, want 0", got)
	}

	wb.SetStepStatus(0, StepDone)
	if got := wb.Progress(); got != 0.25 {
		t.Errorf("Progress = %f, want 0.25", got)
	}

	wb.SetStepStatus(1, StepSkipped) // skipped counts as completed
	wb.SetStepStatus(2, StepDone)
	if got := wb.Progress(); got != 0.75 {
		t.Errorf("Progress = %f, want 0.75", got)
	}

	wb.SetStepStatus(3, StepDone)
	if got := wb.Progress(); got != 1.0 {
		t.Errorf("Progress = %f, want 1.0", got)
	}
}

func TestWorkflowProgressText(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	got := wb.ProgressText()
	if got != "0/2 (0%)" {
		t.Errorf("ProgressText = %q, want '0/2 (0%%)'", got)
	}

	wb.SetStepStatus(0, StepDone)
	got = wb.ProgressText()
	if got != "1/2 (50%)" {
		t.Errorf("ProgressText = %q, want '1/2 (50%%)'", got)
	}
}

func TestWorkflowProgressText_Empty(t *testing.T) {
	wb := NewWorkflowBlock("test")
	got := wb.ProgressText()
	if got != "0/0 (0%)" {
		t.Errorf("ProgressText = %q, want '0/0 (0%%)'", got)
	}
}

func TestWorkflowAutoComplete(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	// Both steps done → block should auto-complete
	wb.SetStepStatus(0, StepDone)
	wb.SetStepStatus(1, StepDone)

	if wb.State() != BlockComplete {
		t.Errorf("State = %v, want BlockComplete", wb.State())
	}
}

func TestWorkflowAutoComplete_WithFailed(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	wb.SetStepStatus(0, StepFailed)
	wb.SetStepStatus(1, StepSkipped)

	if wb.State() != BlockComplete {
		t.Errorf("State = %v, want BlockComplete (failed+skipped are terminal)", wb.State())
	}
}

func TestWorkflowAutoComplete_NotAllTerminal(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	wb.SetStepStatus(0, StepDone)
	// s2 still pending
	if wb.State() == BlockComplete {
		t.Error("should not be complete when a step is still pending")
	}
}

func TestWorkflowHasRunning(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	if wb.HasRunning() {
		t.Error("HasRunning should be false initially")
	}

	wb.SetStepStatus(0, StepRunning)
	if !wb.HasRunning() {
		t.Error("HasRunning should be true after SetStepStatus running")
	}

	wb.SetStepStatus(0, StepDone)
	if wb.HasRunning() {
		t.Error("HasRunning should be false after done")
	}
}

func TestWorkflowHasFailed(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	if wb.HasFailed() {
		t.Error("HasFailed should be false")
	}

	wb.SetStepStatus(0, StepFailed)
	if !wb.HasFailed() {
		t.Error("HasFailed should be true")
	}
}

func TestWorkflowSetTitle(t *testing.T) {
	wb := NewWorkflowBlock("old")
	wb.SetTitle("new title")
	if wb.Title() != "new title" {
		t.Errorf("Title = %q, want 'new title'", wb.Title())
	}
}

func TestWorkflowMeasure(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")
	wb.AddStep("s3", "")

	cs := component.Constraints{MaxWidth: 60}
	size := wb.Measure(cs)
	// Height = title(1) + steps(3) + progress(1) = 5
	if size.H != 5 {
		t.Errorf("Measure H = %d, want 5", size.H)
	}
	if size.W != 60 {
		t.Errorf("Measure W = %d, want 60", size.W)
	}
}

func TestWorkflowMeasure_Empty(t *testing.T) {
	wb := NewWorkflowBlock("test")
	cs := component.Constraints{MaxWidth: 40}
	size := wb.Measure(cs)
	// Height = title(1) + "no steps" placeholder(1) = 2
	if size.H != 2 {
		t.Errorf("Measure H = %d, want 2", size.H)
	}
}

func TestWorkflowMeasure_DefaultWidth(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	cs := component.Constraints{} // no MaxWidth
	size := wb.Measure(cs)
	if size.W != 40 {
		t.Errorf("Measure W = %d, want 40 (default)", size.W)
	}
}

func TestWorkflowPaint(t *testing.T) {
	wb := NewWorkflowBlock("Deploy")
	wb.AddStep("build", "Build project")
	wb.AddStep("test", "Run tests")
	wb.SetStepStatus(0, StepRunning)

	buf := buffer.NewBuffer(60, 10)
	wb.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 5})
	wb.Paint(buf)

	// Title should be rendered
	cell := buf.GetCell(0, 0)
	if cell.Rune == 0 {
		t.Error("title line should have content at (0,0)")
	}

	// First step icon should be at (0, 1)
	iconCell := buf.GetCell(0, 1)
	if iconCell.Rune == 0 {
		t.Error("step icon should be rendered at (0,1)")
	}
}

func TestWorkflowPaint_ProgressBar(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")
	wb.SetStepStatus(0, StepDone)
	wb.SetStepStatus(1, StepDone)

	buf := buffer.NewBuffer(40, 10)
	wb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 5})
	wb.Paint(buf)

	// Progress bar should be at Y=3 (title=0, step0=1, step1=2, bar=3)
	barStart := buf.GetCell(0, 3)
	if barStart.Rune != '[' {
		t.Errorf("bar start = %q, want '['", string(barStart.Rune))
	}
	// Should be fully filled
	filled := buf.GetCell(1, 3)
	if filled.Rune != '█' {
		t.Errorf("bar fill = %q, want '█'", string(filled.Rune))
	}
}

func TestWorkflowPaint_FailedStep(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.SetStepStatus(0, StepFailed)

	buf := buffer.NewBuffer(40, 10)
	wb.SetBounds(component.Rect{X: 0, Y: 0, W: 40, H: 4})
	wb.Paint(buf)

	iconCell := buf.GetCell(0, 1)
	if iconCell.Rune != '✗' {
		t.Errorf("failed icon = %q, want '✗'", string(iconCell.Rune))
	}
	// Red color
	if iconCell.Fg == buffer.NoColor() {
		t.Error("failed icon should have red color")
	}
}

func TestWorkflowSerialize(t *testing.T) {
	wb := NewWorkflowBlock("Deploy")
	wb.AddStep("build", "Build project")
	wb.AddStep("test", "Run tests")
	wb.SetStepStatus(0, StepDone)
	wb.SetStepStatus(1, StepRunning)

	data, err := wb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var raw struct {
		Title string `json:"title"`
		Steps []struct {
			Name   string `json:"name"`
			Status string `json:"status"`
		} `json:"steps"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if raw.Title != "Deploy" {
		t.Errorf("title = %q", raw.Title)
	}
	if len(raw.Steps) != 2 {
		t.Fatalf("steps len = %d", len(raw.Steps))
	}
	if raw.Steps[0].Status != "done" {
		t.Errorf("step[0] status = %q, want 'done'", raw.Steps[0].Status)
	}
	if raw.Steps[1].Status != "running" {
		t.Errorf("step[1] status = %q, want 'running'", raw.Steps[1].Status)
	}
}

func TestWorkflowDeserialize(t *testing.T) {
	jsonData := `{"title":"Loaded","steps":[{"name":"s1","description":"","status":"done","duration_ms":1500},{"name":"s2","description":"desc","status":"pending","duration_ms":0}]}`

	wb := NewWorkflowBlock("temp")
	if err := wb.DeserializeState([]byte(jsonData)); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}

	if wb.Title() != "Loaded" {
		t.Errorf("Title = %q, want 'Loaded'", wb.Title())
	}
	steps := wb.Steps()
	if len(steps) != 2 {
		t.Fatalf("steps len = %d", len(steps))
	}
	if steps[0].Status != StepDone {
		t.Errorf("step[0] status = %v", steps[0].Status)
	}
	if steps[0].Duration != 1500*time.Millisecond {
		t.Errorf("step[0] duration = %v", steps[0].Duration)
	}
	if steps[1].Status != StepPending {
		t.Errorf("step[1] status = %v", steps[1].Status)
	}
}

func TestWorkflowSerializeRoundTrip(t *testing.T) {
	wb := NewWorkflowBlock("Original")
	wb.AddStep("a", "step a")
	wb.AddStep("b", "step b")
	wb.AddStep("c", "step c")
	wb.SetStepStatus(0, StepDone)
	wb.SetStepStatus(1, StepSkipped)
	wb.SetStepStatus(2, StepFailed)

	data, err := wb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	wb2 := NewWorkflowBlock("")
	if err := wb2.DeserializeState(data); err != nil {
		t.Fatalf("DeserializeState: %v", err)
	}
	if wb2.Title() != "Original" {
		t.Errorf("title = %q", wb2.Title())
	}
	steps := wb2.Steps()
	if len(steps) != 3 {
		t.Fatalf("steps len = %d", len(steps))
	}
	if steps[0].Status != StepDone {
		t.Errorf("step[0] = %v", steps[0].Status)
	}
	if steps[1].Status != StepSkipped {
		t.Errorf("step[1] = %v", steps[1].Status)
	}
	if steps[2].Status != StepFailed {
		t.Errorf("step[2] = %v", steps[2].Status)
	}
}

func TestStepStatusString(t *testing.T) {
	cases := []struct {
		status StepStatus
		want   string
	}{
		{StepPending, "pending"},
		{StepRunning, "running"},
		{StepDone, "done"},
		{StepFailed, "failed"},
		{StepSkipped, "skipped"},
	}
	for _, tc := range cases {
		if got := tc.status.String(); got != tc.want {
			t.Errorf("%d.String() = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestStepStatusString_Unknown(t *testing.T) {
	s := StepStatus(99)
	if s.String() != "unknown" {
		t.Errorf("unknown status = %q, want 'unknown'", s.String())
	}
}

func TestStatusIcon(t *testing.T) {
	cases := []struct {
		status StepStatus
		want   rune
	}{
		{StepPending, '○'},
		{StepDone, '✓'},
		{StepFailed, '✗'},
		{StepSkipped, '⊘'},
	}
	for _, tc := range cases {
		if got := statusIcon(tc.status); got != tc.want {
			t.Errorf("statusIcon(%d) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestStatusColor(t *testing.T) {
	// Each status should return a distinct, non-zero color
	colors := make(map[string]bool)
	for _, s := range []StepStatus{StepPending, StepRunning, StepDone, StepFailed, StepSkipped} {
		c := statusColor(s)
		if c == buffer.NoColor() {
			t.Errorf("statusColor(%d) returned NoColor", s)
		}
		colors[c.String()] = true
	}
	// Should have 5 distinct colors
	if len(colors) != 5 {
		t.Errorf("expected 5 distinct colors, got %d", len(colors))
	}
}

func TestWorkflowSpinnerFrame(t *testing.T) {
	wb := NewWorkflowBlock("test")
	r1 := wb.SpinnerFrame()
	if r1 == 0 {
		t.Error("SpinnerFrame should return non-zero rune")
	}
}

func TestWorkflowAdvanceSpinner(t *testing.T) {
	wb := NewWorkflowBlock("test")
	r1 := wb.SpinnerFrame()
	wb.AdvanceSpinner()
	r2 := wb.SpinnerFrame()
	// After advancing, the frame might be the same if not enough time passed,
	// but the internal index should have changed.
	_ = r1
	_ = r2
	// Just verify it doesn't panic and returns valid runes
	if r2 == 0 {
		t.Error("SpinnerFrame should return non-zero rune after advance")
	}
}

func TestWorkflowDurationTracking(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	wb.SetStepStatus(0, StepRunning)
	time.Sleep(5 * time.Millisecond)
	wb.SetStepStatus(0, StepDone)

	steps := wb.Steps()
	if steps[0].Duration < time.Millisecond {
		t.Errorf("Duration = %v, want >= 1ms", steps[0].Duration)
	}
}

func TestWorkflowDurationOnFailed(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	wb.SetStepStatus(0, StepRunning)
	time.Sleep(2 * time.Millisecond)
	wb.SetStepStatus(0, StepFailed)

	steps := wb.Steps()
	if steps[0].Duration <= 0 {
		t.Error("Duration should be > 0 on failed after running")
	}
}

func TestWorkflowReRunningKeepsTimer(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	wb.SetStepStatus(0, StepRunning)
	time.Sleep(5 * time.Millisecond)
	// Setting running again shouldn't reset the timer
	wb.SetStepStatus(0, StepRunning)
	time.Sleep(3 * time.Millisecond)
	wb.SetStepStatus(0, StepDone)

	steps := wb.Steps()
	// Duration should include both sleeps (~8ms total)
	if steps[0].Duration < 5*time.Millisecond {
		t.Errorf("Duration = %v, should include both running periods", steps[0].Duration)
	}
}

func TestWorkflowTypeName(t *testing.T) {
	wb := NewWorkflowBlock("test")
	if wb.TypeName() != "workflow" {
		t.Errorf("TypeName = %q, want 'workflow'", wb.TypeName())
	}
}

func TestWorkflowStepsReturnsCopy(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")
	wb.AddStep("s2", "")

	steps := wb.Steps()
	steps[0].Name = "modified"

	// Original should be unchanged
	origSteps := wb.Steps()
	if origSteps[0].Name != "s1" {
		t.Error("Steps() should return a copy, not a reference")
	}
}

func TestWorkflowPaint_TooSmall(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	// Very small buffer — should not panic
	buf := buffer.NewBuffer(3, 1)
	wb.SetBounds(component.Rect{X: 0, Y: 0, W: 3, H: 1})
	wb.Paint(buf)
}

func TestWorkflowStepIconOverride(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("s1", "")

	// Manually set custom icon
	wb.mu.Lock()
	wb.steps[0].Icon = '⚡'
	wb.mu.Unlock()

	steps := wb.Steps()
	if steps[0].Icon != '⚡' {
		t.Errorf("Icon = %q, want '⚡'", string(steps[0].Icon))
	}
}

func TestWorkflowContainerIntegration(t *testing.T) {
	container := NewBlockContainer()
	wb := NewWorkflowBlock("Agent Workflow")
	wb.AddStep("analyze", "Analyze request")
	wb.AddStep("plan", "Create plan")
	wb.AddStep("execute", "Execute plan")
	wb.SetStepStatus(0, StepDone)
	wb.SetStepStatus(1, StepRunning)

	container.AddBlock(wb)
	if container.Len() != 1 {
		t.Fatalf("container Len = %d, want 1", container.Len())
	}

	block := container.Blocks()[0]
	wf, ok := block.(*WorkflowBlock)
	if !ok {
		t.Fatalf("expected *WorkflowBlock, got %T", block)
	}
	if wf.Title() != "Agent Workflow" {
		t.Errorf("Title = %q", wf.Title())
	}
	if wf.StepCount() != 3 {
		t.Errorf("StepCount = %d, want 3", wf.StepCount())
	}
}

func TestWorkflowRealisticScenario(t *testing.T) {
	wb := NewWorkflowBlock("Code Review Agent")
	wb.AddStep("fetch-diff", "Fetch PR diff")
	wb.AddStep("analyze", "Analyze changes")
	wb.AddStep("review", "Generate review comments")
	wb.AddStep("submit", "Submit review")

	// Verify all pending
	if got := wb.Progress(); got != 0 {
		t.Errorf("initial Progress = %f, want 0", got)
	}
	if wb.HasRunning() {
		t.Error("should not have running initially")
	}

	// Start first step
	wb.SetStepStatus(0, StepRunning)
	if !wb.HasRunning() {
		t.Error("should have running after SetStepStatus(0, Running)")
	}
	if wb.State() == BlockComplete {
		t.Error("should not be complete yet")
	}

	// Complete step 0
	wb.SetStepStatus(0, StepDone)
	// Skip step 1 (conditional)
	wb.SetStepStatus(1, StepSkipped)
	// Step 2 fails
	wb.SetStepStatus(2, StepFailed)

	// 3/4 terminal, step 3 still pending
	if wb.State() == BlockComplete {
		t.Error("should not be complete with step 3 pending")
	}

	// Complete step 3
	wb.SetStepStatus(3, StepDone)
	if wb.State() != BlockComplete {
		t.Error("should be complete after all steps terminal")
	}

	// Progress should be 3/4 (done + skipped = completed, failed = terminal but not completed)
	progress := wb.Progress()
	// done=2 (s0, s3), skipped=1 (s1), failed=1 (s2)
	// completed (done+skipped) = 3, total = 4 → 0.75
	if progress != 0.75 {
		t.Errorf("Progress = %f, want 0.75", progress)
	}

	if !wb.HasFailed() {
		t.Error("should have failed step")
	}
}

func TestFormatDurationWf(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{500 * time.Microsecond, "<1ms"},
		{5 * time.Millisecond, "5ms"},
		{1500 * time.Millisecond, "1.5s"},
		{65 * time.Second, "1m5s"},
	}
	for _, tc := range cases {
		got := formatDurationWf(tc.d)
		if got != tc.want {
			t.Errorf("formatDuration(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestWorkflowPaint_WithDescription(t *testing.T) {
	wb := NewWorkflowBlock("test")
	wb.AddStep("build", "Build the project")
	wb.SetStepStatus(0, StepDone)

	buf := buffer.NewBuffer(60, 10)
	wb.SetBounds(component.Rect{X: 0, Y: 0, W: 60, H: 4})
	wb.Paint(buf)

	// The step name should be rendered
	nameCell := buf.GetCell(2, 1) // x+2, y+1
	if nameCell.Rune != 'b' {
		t.Errorf("step name char = %q, want 'b'", string(nameCell.Rune))
	}
}

func TestWorkflowEmptyStepsSerialize(t *testing.T) {
	wb := NewWorkflowBlock("empty")
	data, err := wb.SerializeState()
	if err != nil {
		t.Fatalf("SerializeState: %v", err)
	}

	var raw struct {
		Title string `json:"title"`
		Steps []struct{} `json:"steps"`
	}
	json.Unmarshal(data, &raw)
	if raw.Title != "empty" {
		t.Errorf("title = %q", raw.Title)
	}
	if len(raw.Steps) != 0 {
		t.Errorf("steps len = %d, want 0", len(raw.Steps))
	}
}
