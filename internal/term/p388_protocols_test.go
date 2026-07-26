package term

import "testing"

// P388: Tests for OSC 7 (CWD), OSC 777 (URxvt notification), OSC 9;4 (progress bar)

func TestP388_ReportWorkingDir(t *testing.T) {
	got := ReportWorkingDir("/home/user/project")
	want := "\x1b]7;file:///home/user/project\x07"
	if got != want {
		t.Errorf("ReportWorkingDir() = %q, want %q", got, want)
	}
}

func TestP388_ReportWorkingDirHost(t *testing.T) {
	got := ReportWorkingDirHost("localhost", "/var/log")
	want := "\x1b]7;file://localhost/var/log\x07"
	if got != want {
		t.Errorf("ReportWorkingDirHost() = %q, want %q", got, want)
	}
}

func TestP388_URxvtNotification(t *testing.T) {
	got := URxvtNotification("Build Complete", "All tests passed")
	if got[:5] != "\x1b]777" {
		t.Errorf("URxvtNotification prefix = %q", got[:5])
	}
	if got[len(got)-1] != 0x07 {
		t.Error("URxvtNotification should end with BEL")
	}
}

func TestP388_URxvtNotification_Empty(t *testing.T) {
	got := URxvtNotification("", "")
	if got != "\x1b]777;notify;;\x07" {
		t.Errorf("URxvtNotification empty = %q", got)
	}
}

func TestP388_SetTabProgressBar_Set(t *testing.T) {
	got := SetTabProgressBar(50, ProgressBarSet)
	want := "\x1b]9;4;;50\x07"
	if got != want {
		t.Errorf("SetTabProgressBar(50, Set) = %q, want %q", got, want)
	}
}

func TestP388_SetTabProgressBar_Error(t *testing.T) {
	got := SetTabProgressBar(75, ProgressBarError)
	want := "\x1b]9;4;2;75\x07"
	if got != want {
		t.Errorf("Error progress = %q, want %q", got, want)
	}
}

func TestP388_SetTabProgressBar_Warning(t *testing.T) {
	got := SetTabProgressBar(30, ProgressBarWarning)
	want := "\x1b]9;4;3;30\x07"
	if got != want {
		t.Errorf("Warning progress = %q, want %q", got, want)
	}
}

func TestP388_SetTabProgressBar_Indeterminate(t *testing.T) {
	got := SetTabProgressBar(0, ProgressBarIndeterminate)
	want := "\x1b]9;4;4;0\x07"
	if got != want {
		t.Errorf("Indeterminate progress = %q, want %q", got, want)
	}
}

func TestP388_SetTabProgressBar_Clamp(t *testing.T) {
	// Negative clamps to 0
	got := SetTabProgressBar(-50, ProgressBarSet)
	if got != "\x1b]9;4;;0\x07" {
		t.Errorf("Negative clamp = %q", got)
	}
	// >100 clamps to 100
	got = SetTabProgressBar(150, ProgressBarSet)
	if got != "\x1b]9;4;;100\x07" {
		t.Errorf("Over-100 clamp = %q", got)
	}
}

func TestP388_SetTabProgressBar_Clear(t *testing.T) {
	got := SetTabProgressBar(0, ProgressBarClear)
	want := "\x1b]9;4;1;0\x07"
	if got != want {
		t.Errorf("Clear progress = %q, want %q", got, want)
	}
}

func TestP388_ClearTabProgressBar(t *testing.T) {
	got := ClearTabProgressBar()
	want := "\x1b]9;4;0;\x07"
	if got != want {
		t.Errorf("ClearTabProgressBar() = %q, want %q", got, want)
	}
}

func TestP388_EnableFocusReporting(t *testing.T) {
	got := EnableFocusReporting
	want := "\x1b[?1004h"
	if got != want {
		t.Errorf("EnableFocusReporting = %q, want %q", got, want)
	}
}

func TestP388_DisableFocusReporting(t *testing.T) {
	got := DisableFocusReporting
	want := "\x1b[?1004l"
	if got != want {
		t.Errorf("DisableFocusReporting = %q, want %q", got, want)
	}
}
