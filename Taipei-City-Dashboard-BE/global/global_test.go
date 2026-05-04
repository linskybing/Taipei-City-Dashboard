package global

import "testing"

func TestGetBoundedIntEnvMemoryTurns(t *testing.T) {
	tests := []struct {
		name string
		set  bool
		raw  string
		want int
	}{
		{name: "unset", want: 8},
		{name: "below", set: true, raw: "0", want: 4},
		{name: "above", set: true, raw: "999", want: 12},
		{name: "invalid", set: true, raw: "abc", want: 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_TWCC_MEMORY_TURNS_" + tt.name
			if tt.set {
				t.Setenv(key, tt.raw)
			}
			got := getBoundedIntEnv(key, 8, 4, 12)
			if got != tt.want {
				t.Fatalf("getBoundedIntEnv() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetBoundedIntEnvMemoryBlockRunes(t *testing.T) {
	tests := []struct {
		name string
		set  bool
		raw  string
		want int
	}{
		{name: "unset", want: 1600},
		{name: "below", set: true, raw: "100", want: 900},
		{name: "above", set: true, raw: "99999", want: 2400},
		{name: "invalid", set: true, raw: "abc", want: 1600},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := "TEST_TWCC_MEMORY_BLOCK_RUNES_" + tt.name
			if tt.set {
				t.Setenv(key, tt.raw)
			}
			got := getBoundedIntEnv(key, 1600, 900, 2400)
			if got != tt.want {
				t.Fatalf("getBoundedIntEnv() = %d, want %d", got, tt.want)
			}
		})
	}
}
