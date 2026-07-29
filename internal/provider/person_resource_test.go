package provider

import (
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestParseBirthDate(t *testing.T) {
	tests := []struct {
		input    string
		expected string // YYYY-MM-DD
		valid    bool
	}{
		{"2023-06-10", "2023-06-10", true},
		{"1988-05-16", "1988-05-16", true},
		{"2023-06-10T00:00:00.000Z", "2023-06-10", true},
		{"2023-06-10T04:00:00Z", "2023-06-10", true},
		{"2023-06-10T12:30:45Z", "2023-06-10", true},
		{"2023-06-10 00:00:00", "2023-06-10", true},
		{"invalid-date", "", false},
		{"", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			parsed, ok := parseBirthDate(tt.input)
			if ok != tt.valid {
				t.Fatalf("expected valid=%v for %q, got %v", tt.valid, tt.input, ok)
			}
			if ok {
				formatted := parsed.Format("2006-01-02")
				if formatted != tt.expected {
					t.Errorf("expected %q, got %q", tt.expected, formatted)
				}
				if parsed.Location() != time.UTC {
					t.Errorf("expected UTC location, got %v", parsed.Location())
				}
			}
		})
	}
}

func TestReconcileBirthDate(t *testing.T) {
	strPtr := func(s string) *string {
		return &s
	}

	tests := []struct {
		name       string
		apiVal     *string
		currentVal types.String
		expected   types.String
	}{
		{
			name:       "nil API value",
			apiVal:     nil,
			currentVal: types.StringValue("2023-06-10"),
			expected:   types.StringNull(),
		},
		{
			name:       "empty API string",
			apiVal:     strPtr(""),
			currentVal: types.StringValue("2023-06-10"),
			expected:   types.StringNull(),
		},
		{
			name:       "ISO timestamp from API matches YYYY-MM-DD in state",
			apiVal:     strPtr("2023-06-10T00:00:00.000Z"),
			currentVal: types.StringValue("2023-06-10"),
			expected:   types.StringValue("2023-06-10"),
		},
		{
			name:       "ISO timestamp from API with timezone offset matches YYYY-MM-DD in state",
			apiVal:     strPtr("1988-05-16T00:00:00.000Z"),
			currentVal: types.StringValue("1988-05-16"),
			expected:   types.StringValue("1988-05-16"),
		},
		{
			name:       "ISO timestamp from API when state is null",
			apiVal:     strPtr("2023-06-10T00:00:00.000Z"),
			currentVal: types.StringNull(),
			expected:   types.StringValue("2023-06-10"),
		},
		{
			name:       "Date changes in API",
			apiVal:     strPtr("2023-06-11T00:00:00.000Z"),
			currentVal: types.StringValue("2023-06-10"),
			expected:   types.StringValue("2023-06-11"),
		},
		{
			name:       "Exact ISO format in state preserved if matching",
			apiVal:     strPtr("2023-06-10T00:00:00.000Z"),
			currentVal: types.StringValue("2023-06-10T00:00:00Z"),
			expected:   types.StringValue("2023-06-10T00:00:00Z"),
		},
		{
			name:       "ISO timestamp from API shifted to previous UTC day (e.g. 19:00Z) matches YYYY-MM-DD in state",
			apiVal:     strPtr("1988-03-03T19:00:00.000Z"),
			currentVal: types.StringValue("1988-03-04"),
			expected:   types.StringValue("1988-03-04"),
		},
		{
			name:       "ISO timestamp from API shifted to previous UTC day (e.g. 23:00Z) matches YYYY-MM-DD in state",
			apiVal:     strPtr("1988-03-03T23:00:00.000Z"),
			currentVal: types.StringValue("1988-03-04"),
			expected:   types.StringValue("1988-03-04"),
		},
		{
			name:       "ISO timestamp from API shifted into positive UTC hours (e.g. 05:00Z) matches YYYY-MM-DD in state",
			apiVal:     strPtr("1988-03-04T05:00:00.000Z"),
			currentVal: types.StringValue("1988-03-04"),
			expected:   types.StringValue("1988-03-04"),
		},
		{
			name:       "ISO timestamp from API shifted to previous UTC day (e.g. 19:00Z) when state is null resolves to target date",
			apiVal:     strPtr("1988-03-03T19:00:00.000Z"),
			currentVal: types.StringNull(),
			expected:   types.StringValue("1988-03-04"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reconcileBirthDate(tt.apiVal, tt.currentVal)
			if result.IsNull() != tt.expected.IsNull() {
				t.Fatalf("expected IsNull=%v, got %v", tt.expected.IsNull(), result.IsNull())
			}
			if !result.IsNull() && result.ValueString() != tt.expected.ValueString() {
				t.Errorf("expected %q, got %q", tt.expected.ValueString(), result.ValueString())
			}
		})
	}
}
