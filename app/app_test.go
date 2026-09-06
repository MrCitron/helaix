package main

import "testing"

func TestValidateExportFilename(t *testing.T) {
	tests := []struct {
		filename string
		want     string
		valid    bool
	}{
		{filename: "My Preset.hlx", want: "My Preset", valid: true},
		{filename: "My Preset", want: "My Preset", valid: true},
		{filename: "../outside.hlx"},
		{filename: "nested/preset.hlx"},
		{filename: `nested\\preset.hlx`},
		{filename: "preset.json"},
		{filename: ""},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got, err := validateExportFilename(tt.filename)
			if tt.valid && (err != nil || got != tt.want) {
				t.Fatalf("validateExportFilename() = %q, %v; want %q, nil", got, err, tt.want)
			}
			if !tt.valid && err == nil {
				t.Fatalf("validateExportFilename() accepted invalid filename %q", tt.filename)
			}
		})
	}
}
