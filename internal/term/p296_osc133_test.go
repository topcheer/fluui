package term

import "testing"

func TestP296_OSC133Sequences(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{"PromptStart", OSC133PromptStartSeq(), "\x1b]133;A\x07"},
		{"CommandStart", OSC133CommandStartSeq(), "\x1b]133;B\x07"},
		{"OutputStart", OSC133OutputStartSeq(), "\x1b]133;C\x07"},
		{"CommandEnd_0", OSC133CommandEndSeq(0), "\x1b]133;D;0\x07"},
		{"CommandEnd_1", OSC133CommandEndSeq(1), "\x1b]133;D;1\x07"},
		{"CommandEnd_127", OSC133CommandEndSeq(127), "\x1b]133;D;127\x07"},
		{"PromptStartMeta_empty", OSC133PromptStartMeta(""), "\x1b]133;A\x07"},
		{"PromptStartMeta", OSC133PromptStartMeta("tmux=prompt"), "\x1b]133;A;tmux=prompt\x07"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestP296_ParseOSC133(t *testing.T) {
	tests := []struct {
		name   string
		input  string
		mark   OSC133Mark
		exit   int
		meta   string
		ok     bool
	}{
		{"prompt_start_bel", "\x1b]133;A\x07", OSC133PromptStart, 0, "", true},
		{"prompt_start_st", "\x1b]133;A\x1b\\", OSC133PromptStart, 0, "", true},
		{"prompt_start_meta", "\x1b]133;A;claude=1\x07", OSC133PromptStart, 0, "claude=1", true},
		{"command_start", "\x1b]133;B\x07", OSC133CommandStart, 0, "", true},
		{"output_start", "\x1b]133;C\x07", OSC133OutputStart, 0, "", true},
		{"command_end_0", "\x1b]133;D;0\x07", OSC133CommandEnd, 0, "", true},
		{"command_end_1", "\x1b]133;D;1\x07", OSC133CommandEnd, 1, "", true},
		{"command_end_no_code", "\x1b]133;D\x07", OSC133CommandEnd, -1, "", true},
		{"command_end_st", "\x1b]133;D;42\x1b\\", OSC133CommandEnd, 42, "", true},
		// failures
		{"empty", "", OSC133Unknown, 0, "", false},
		{"too_short", "\x1b]1\x07", OSC133Unknown, 0, "", false},
		{"wrong_osc", "\x1b]134;A\x07", OSC133Unknown, 0, "", false},
		{"unknown_mark", "\x1b]133;X\x07", OSC133Unknown, 0, "", false},
		{"no_terminator", "\x1b]133;A", OSC133Unknown, 0, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, ok := ParseOSC133(tt.input)
			if ok != tt.ok {
				t.Errorf("ok = %v, want %v", ok, tt.ok)
				return
			}
			if ok {
				if result.Mark != tt.mark {
					t.Errorf("mark = %v, want %v", result.Mark, tt.mark)
				}
				if result.ExitCode != tt.exit {
					t.Errorf("exit = %d, want %d", result.ExitCode, tt.exit)
				}
				if result.Meta != tt.meta {
					t.Errorf("meta = %q, want %q", result.Meta, tt.meta)
				}
			}
		})
	}
}

func TestP296_OSC133FullCycle(t *testing.T) {
	// Simulate a full prompt→command→output→end cycle
	seqs := []string{
		OSC133PromptStartSeq(),
		OSC133CommandStartSeq(),
		OSC133OutputStartSeq(),
		OSC133CommandEndSeq(0),
	}
	expected := []OSC133Mark{
		OSC133PromptStart,
		OSC133CommandStart,
		OSC133OutputStart,
		OSC133CommandEnd,
	}
	for i, s := range seqs {
		result, ok := ParseOSC133(s)
		if !ok {
			t.Errorf("seq %d: parse failed for %q", i, s)
			continue
		}
		if result.Mark != expected[i] {
			t.Errorf("seq %d: mark = %v, want %v", i, result.Mark, expected[i])
		}
	}
}
