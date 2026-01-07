package theme

import "testing"

func TestCatppuccinMocha(t *testing.T) {
	theme := CatppuccinMocha()
	if theme == nil {
		t.Fatal("CatppuccinMocha() returned nil")
	}

	tests := []struct {
		name  string
		value string
	}{
		{"Background", "#1E1E2E"},
		{"Primary", "#89dceb"},
		{"Secondary", "#cba6f7"},
		{"Muted", "#6c7086"},
		{"Success", "#a6e3a1"},
		{"Warning", "#fab387"},
		{"Error", "#f38ba8"},
		{"Info", "#b4befe"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got string
			switch tt.name {
			case "Background":
				got = theme.Background
			case "Primary":
				got = theme.Primary
			case "Secondary":
				got = theme.Secondary
			case "Muted":
				got = theme.Muted
			case "Success":
				got = theme.Success
			case "Warning":
				got = theme.Warning
			case "Error":
				got = theme.Error
			case "Info":
				got = theme.Info
			}

			if got != tt.value {
				t.Errorf("%s = %s, want %s", tt.name, got, tt.value)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	theme := Default()
	if theme == nil {
		t.Fatal("Default() returned nil")
	}
	// Default should be Catppuccin Mocha
	if theme.Background != "#1E1E2E" {
		t.Errorf("Default() Background = %s, want #1E1E2E", theme.Background)
	}
}

func TestColorNames(t *testing.T) {
	colors := ColorNames()
	if colors == nil {
		t.Fatal("ColorNames() returned nil")
	}

	expectedKeys := []string{
		"background", "primary", "secondary", "muted",
		"success", "warning", "error", "info",
	}

	for _, key := range expectedKeys {
		if colors[key] == "" {
			t.Errorf("ColorNames() missing key: %s", key)
		}
	}
}

func TestANSIColors(t *testing.T) {
	colors := ANSIColors()
	if colors == nil {
		t.Fatal("ANSIColors() returned nil")
	}

	expectedKeys := []string{
		"primary", "secondary", "muted",
		"success", "warning", "error", "info",
	}

	for _, key := range expectedKeys {
		if _, ok := colors[key]; !ok {
			t.Errorf("ANSIColors() missing key: %s", key)
		}
	}

	// Verify some known values
	if colors["primary"] != 38 {
		t.Errorf("ANSIColors()[primary] = %d, want 38", colors["primary"])
	}
	if colors["error"] != 203 {
		t.Errorf("ANSIColors()[error] = %d, want 203", colors["error"])
	}
}
