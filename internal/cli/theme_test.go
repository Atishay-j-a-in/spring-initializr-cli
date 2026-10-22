package cli

import "testing"

func TestThemeStyling(t *testing.T) {
	if got := (Theme{}).Accent("text"); got != "text" {
		t.Fatalf("plain theme changed text: %q", got)
	}
	if got := (Theme{Enabled: true}).Success("done"); got != ansiBold+ansiGreen+"done"+ansiReset {
		t.Fatalf("unexpected styled text: %q", got)
	}
}
