package model

import (
	"testing"
)

// FuzzNormalizeTags tests tag normalization with random inputs.
// Run with: go test -fuzz=FuzzNormalizeTags -fuzztime=30s ./model/...
func FuzzNormalizeTags(f *testing.F) {
	// Seed corpus with valid tags
	f.Add("valid-tag")
	f.Add("tag_with_underscore")
	f.Add("tag.with.dots")
	f.Add("TAG123")
	f.Add("MixedCase")

	// Whitespace variations
	f.Add("  spaces  ")
	f.Add("multiple   spaces")
	f.Add("\ttabs\t")
	f.Add("  leading")
	f.Add("trailing  ")

	// Special characters (should be filtered)
	f.Add("tag@special")
	f.Add("tag#hash")
	f.Add("tag$dollar")
	f.Add("tag!exclaim")
	f.Add("@#$%^&*()")

	// Edge cases
	f.Add("")
	f.Add("   ")
	f.Add("a")
	f.Add("verylongtagnamethatmightexceedthelimitifitkeepsgoingandgoingandgoing")

	f.Fuzz(func(t *testing.T, input string) {
		tags := []string{input}

		// NormalizeTags should not panic
		result := NormalizeTags(tags)

		// Result should have at most 20 tags
		if len(result) > 20 {
			t.Errorf("NormalizeTags returned %d tags, max is 20", len(result))
		}

		// Each result tag should be valid
		for _, tag := range result {
			// Should be lowercase
			for _, c := range tag {
				if c >= 'A' && c <= 'Z' {
					t.Errorf("NormalizeTags returned uppercase character in %q", tag)
				}
			}

			// Should only contain allowed characters
			for _, c := range tag {
				valid := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.'
				if !valid {
					t.Errorf("NormalizeTags returned invalid character %q in %q", string(c), tag)
				}
			}

			// Should not be empty
			if tag == "" {
				t.Error("NormalizeTags returned empty tag")
			}

			// Should not exceed max length
			if len(tag) > 64 {
				t.Errorf("NormalizeTags returned tag longer than 64 chars: %d", len(tag))
			}
		}

		// Should be deduplicated
		seen := make(map[string]bool)
		for _, tag := range result {
			if seen[tag] {
				t.Errorf("NormalizeTags returned duplicate tag: %q", tag)
			}
			seen[tag] = true
		}
	})
}

// FuzzNormalizeTagsMultiple tests tag normalization with multiple tags.
// Run with: go test -fuzz=FuzzNormalizeTagsMultiple -fuzztime=30s ./model/...
func FuzzNormalizeTagsMultiple(f *testing.F) {
	// Seed with multiple tag combinations
	f.Add("tag1", "tag2", "tag3")
	f.Add("duplicate", "duplicate", "unique")
	f.Add("UPPER", "lower", "MiXeD")
	f.Add("", "valid", "")
	f.Add("a b c", "d-e-f", "g_h_i")

	f.Fuzz(func(t *testing.T, a, b, c string) {
		tags := []string{a, b, c}

		// NormalizeTags should not panic
		result := NormalizeTags(tags)

		// Result should have at most 20 tags (and at most 3 for this input)
		if len(result) > 20 {
			t.Errorf("NormalizeTags returned %d tags, max is 20", len(result))
		}

		// Verify deduplication
		seen := make(map[string]bool)
		for _, tag := range result {
			if seen[tag] {
				t.Errorf("NormalizeTags returned duplicate tag: %q", tag)
			}
			seen[tag] = true
		}
	})
}

// FuzzParseToolID tests tool ID parsing with random inputs.
// Run with: go test -fuzz=FuzzParseToolID -fuzztime=30s ./model/...
func FuzzParseToolID(f *testing.F) {
	// Valid tool IDs
	f.Add("namespace:name")
	f.Add("simple-name")
	f.Add("ns:tool-name")
	f.Add("my-namespace:my-tool")

	// Invalid tool IDs
	f.Add("")
	f.Add(":")
	f.Add(":name")
	f.Add("namespace:")
	f.Add("a:b:c")
	f.Add("too:many:colons")

	// Edge cases
	f.Add("a")
	f.Add("a:b")
	f.Add("-")
	f.Add("_")

	f.Fuzz(func(t *testing.T, input string) {
		// ParseToolID should not panic
		namespace, name, err := ParseToolID(input)
		if err != nil {
			// Invalid input is expected
			return
		}

		// If successful, both should follow rules
		if input == "" {
			t.Error("ParseToolID succeeded on empty input")
		}

		// If there was a colon, both namespace and name should be non-empty
		hasColon := false
		for _, c := range input {
			if c == ':' {
				hasColon = true
				break
			}
		}

		if hasColon {
			if namespace == "" || name == "" {
				t.Errorf("ParseToolID(%q) returned empty namespace=%q or name=%q with colon", input, namespace, name)
			}
		} else {
			if namespace != "" {
				t.Errorf("ParseToolID(%q) returned namespace=%q without colon", input, namespace)
			}
			if name != input {
				t.Errorf("ParseToolID(%q) name=%q doesn't match input", input, name)
			}
		}
	})
}

// FuzzSchemaValidation tests schema validation with random JSON-like structures.
// Run with: go test -fuzz=FuzzSchemaValidation -fuzztime=30s ./model/...
func FuzzSchemaValidation(f *testing.F) {
	// Valid schemas
	f.Add(`{"type":"object"}`)
	f.Add(`{"type":"object","properties":{"name":{"type":"string"}}}`)
	f.Add(`{"type":"object","required":["name"]}`)
	f.Add(`{"type":"string"}`)
	f.Add(`{"type":"number"}`)
	f.Add(`{"type":"boolean"}`)
	f.Add(`{"type":"array","items":{"type":"string"}}`)

	// Invalid schemas
	f.Add(`{}`)
	f.Add(`{"type":"invalid"}`)
	f.Add(`not json`)
	f.Add(``)
	f.Add(`null`)

	// Edge cases
	f.Add(`{"type":"object","additionalProperties":false}`)
	f.Add(`{"$schema":"https://json-schema.org/draft/2020-12/schema","type":"object"}`)

	f.Fuzz(func(t *testing.T, schemaJSON string) {
		v := NewDefaultValidator()

		// Try to validate empty input against the schema
		// This should not panic regardless of schema validity
		_ = v.Validate([]byte(schemaJSON), map[string]any{})

		// Also try with json.RawMessage
		_ = v.Validate([]byte(schemaJSON), nil)
	})
}
