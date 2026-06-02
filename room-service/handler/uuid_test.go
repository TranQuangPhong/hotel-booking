package handler

import (
	"regexp"
	"testing"
)

func TestGenerateUUID_Format(t *testing.T) {
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)

	for i := 0; i < 100; i++ {
		id := generateUUID()
		if !uuidRegex.MatchString(id) {
			t.Fatalf("generateUUID() = %q, does not match UUID v4 format", id)
		}
	}
}

func TestGenerateUUID_Uniqueness(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := generateUUID()
		if seen[id] {
			t.Fatalf("generateUUID() produced duplicate: %s", id)
		}
		seen[id] = true
	}
}
