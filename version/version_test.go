package version

import (
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		input   string
		want    Version
		wantErr bool
	}{
		{"1.0.0", Version{1, 0, 0, "", ""}, false},
		{"v1.0.0", Version{1, 0, 0, "", ""}, false},
		{"1.2.3", Version{1, 2, 3, "", ""}, false},
		{"1.0.0-alpha", Version{1, 0, 0, "alpha", ""}, false},
		{"1.0.0-alpha.1", Version{1, 0, 0, "alpha.1", ""}, false},
		{"1.0.0+build", Version{1, 0, 0, "", "build"}, false},
		{"1.0.0-beta+build.123", Version{1, 0, 0, "beta", "build.123"}, false},
		{"invalid", Version{}, true},
		{"1.0", Version{}, true},
		{"1", Version{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := Parse(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestMustParse(t *testing.T) {
	// Test valid parse
	v := MustParse("1.2.3")
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 {
		t.Errorf("MustParse failed: got %v", v)
	}

	// Test panic on invalid
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("MustParse did not panic on invalid input")
		}
	}()
	MustParse("invalid")
}

func TestVersion_String(t *testing.T) {
	tests := []struct {
		v    Version
		want string
	}{
		{Version{1, 0, 0, "", ""}, "v1.0.0"},
		{Version{1, 2, 3, "", ""}, "v1.2.3"},
		{Version{1, 0, 0, "alpha", ""}, "v1.0.0-alpha"},
		{Version{1, 0, 0, "", "build"}, "v1.0.0+build"},
		{Version{1, 0, 0, "beta", "123"}, "v1.0.0-beta+123"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.v.String(); got != tt.want {
				t.Errorf("Version.String() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestVersion_Compare(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.0", "1.0.0", 0},
		{"1.0.0", "2.0.0", -1},
		{"2.0.0", "1.0.0", 1},
		{"1.0.0", "1.1.0", -1},
		{"1.1.0", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.0-alpha", "1.0.0", -1},
		{"1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha", "1.0.0-beta", -1},
	}

	for _, tt := range tests {
		t.Run(tt.a+"_vs_"+tt.b, func(t *testing.T) {
			a := MustParse(tt.a)
			b := MustParse(tt.b)
			if got := a.Compare(b); got != tt.want {
				t.Errorf("%s.Compare(%s) = %d, want %d", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestVersion_LessThan(t *testing.T) {
	if !MustParse("1.0.0").LessThan(MustParse("2.0.0")) {
		t.Error("1.0.0 should be less than 2.0.0")
	}
	if MustParse("2.0.0").LessThan(MustParse("1.0.0")) {
		t.Error("2.0.0 should not be less than 1.0.0")
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	if !MustParse("2.0.0").GreaterThan(MustParse("1.0.0")) {
		t.Error("2.0.0 should be greater than 1.0.0")
	}
	if MustParse("1.0.0").GreaterThan(MustParse("2.0.0")) {
		t.Error("1.0.0 should not be greater than 2.0.0")
	}
}

func TestVersion_Equal(t *testing.T) {
	if !MustParse("1.0.0").Equal(MustParse("1.0.0")) {
		t.Error("1.0.0 should equal 1.0.0")
	}
	if MustParse("1.0.0").Equal(MustParse("1.0.1")) {
		t.Error("1.0.0 should not equal 1.0.1")
	}
}

func TestVersion_Compatible(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"1.0.0", "1.0.0", true},
		{"1.1.0", "1.0.0", true},
		{"1.0.0", "1.1.0", false}, // v < other
		{"2.0.0", "1.0.0", false}, // different major
		{"1.0.0", "2.0.0", false}, // different major
	}

	for _, tt := range tests {
		t.Run(tt.a+"_compat_"+tt.b, func(t *testing.T) {
			a := MustParse(tt.a)
			b := MustParse(tt.b)
			if got := a.Compatible(b); got != tt.want {
				t.Errorf("%s.Compatible(%s) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
		})
	}
}

func TestConstraint_Check(t *testing.T) {
	tests := []struct {
		constraint string
		version    string
		want       bool
	}{
		{"1.0.0", "1.0.0", true},
		{"=1.0.0", "1.0.0", true},
		{"=1.0.0", "1.0.1", false},
		{">1.0.0", "1.0.1", true},
		{">1.0.0", "1.0.0", false},
		{">=1.0.0", "1.0.0", true},
		{">=1.0.0", "0.9.0", false},
		{"<2.0.0", "1.9.9", true},
		{"<2.0.0", "2.0.0", false},
		{"<=2.0.0", "2.0.0", true},
		{"^1.0.0", "1.5.0", true},
		{"^1.0.0", "2.0.0", false},
		{"~1.0.0", "1.0.5", true},
		{"~1.0.0", "1.1.0", false},
	}

	for _, tt := range tests {
		t.Run(tt.constraint+"_"+tt.version, func(t *testing.T) {
			c, err := ParseConstraint(tt.constraint)
			if err != nil {
				t.Fatalf("ParseConstraint(%q) error: %v", tt.constraint, err)
			}
			v := MustParse(tt.version)
			if got := c.Check(v); got != tt.want {
				t.Errorf("Constraint(%q).Check(%s) = %v, want %v", tt.constraint, tt.version, got, tt.want)
			}
		})
	}
}

func TestConstraint_String(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"1.0.0", "v1.0.0"},
		{"=1.0.0", "v1.0.0"},
		{">1.0.0", ">v1.0.0"},
		{">=1.0.0", ">=v1.0.0"},
		{"^1.0.0", "^v1.0.0"},
		{"~1.0.0", "~v1.0.0"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			c, _ := ParseConstraint(tt.input)
			if got := c.String(); got != tt.want {
				t.Errorf("Constraint(%q).String() = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseConstraint_Invalid(t *testing.T) {
	_, err := ParseConstraint("invalid")
	if err == nil {
		t.Error("ParseConstraint should fail on invalid version")
	}
}

func TestMatrix_Check(t *testing.T) {
	m := NewMatrix()
	m.Add(Compatibility{
		Component:  "test",
		MinVersion: MustParse("1.0.0"),
	})

	// Test compatible version
	ok, msg := m.Check("test", MustParse("1.5.0"))
	if !ok {
		t.Errorf("Expected compatible, got: %s", msg)
	}

	// Test version below minimum
	ok, _ = m.Check("test", MustParse("0.5.0"))
	if ok {
		t.Error("Expected incompatible for version below minimum")
	}

	// Test unknown component (should be compatible)
	ok, _ = m.Check("unknown", MustParse("1.0.0"))
	if !ok {
		t.Error("Unknown component should be compatible")
	}
}

func TestMatrix_Check_MaxVersion(t *testing.T) {
	m := NewMatrix()
	maxV := MustParse("2.0.0")
	m.Add(Compatibility{
		Component:  "test",
		MinVersion: MustParse("1.0.0"),
		MaxVersion: &maxV,
	})

	// Test within range
	ok, _ := m.Check("test", MustParse("1.5.0"))
	if !ok {
		t.Error("Version within range should be compatible")
	}

	// Test above max
	ok, _ = m.Check("test", MustParse("2.5.0"))
	if ok {
		t.Error("Version above max should be incompatible")
	}
}

func TestMatrix_Check_Deprecated(t *testing.T) {
	m := NewMatrix()
	m.Add(Compatibility{
		Component:  "test",
		MinVersion: MustParse("1.0.0"),
		Deprecated: true,
		Message:    "Use v2 instead",
	})

	ok, msg := m.Check("test", MustParse("1.5.0"))
	if !ok {
		t.Error("Deprecated should still be compatible")
	}
	if msg != "Use v2 instead" {
		t.Errorf("Expected deprecation message, got: %s", msg)
	}
}

func TestMatrix_Negotiate(t *testing.T) {
	m := NewMatrix()
	m.Add(Compatibility{
		Component:  "test",
		MinVersion: MustParse("1.0.0"),
	})

	available := []Version{
		MustParse("0.5.0"),
		MustParse("1.0.0"),
		MustParse("1.5.0"),
		MustParse("2.0.0"),
	}

	best, err := m.Negotiate("test", available)
	if err != nil {
		t.Fatalf("Negotiate failed: %v", err)
	}
	if best.String() != "v2.0.0" {
		t.Errorf("Expected v2.0.0, got %s", best.String())
	}
}

func TestMatrix_Negotiate_NoCompatible(t *testing.T) {
	m := NewMatrix()
	m.Add(Compatibility{
		Component:  "test",
		MinVersion: MustParse("2.0.0"),
	})

	available := []Version{
		MustParse("0.5.0"),
		MustParse("1.0.0"),
	}

	_, err := m.Negotiate("test", available)
	if err == nil {
		t.Error("Negotiate should fail when no compatible version exists")
	}
}

func TestConstraint_Check_InvalidOperator(t *testing.T) {
	// Create a constraint with an invalid operator to hit the default case
	c := Constraint{
		Op:      "!!",
		Version: MustParse("1.0.0"),
	}
	v := MustParse("1.0.0")
	if c.Check(v) {
		t.Error("Constraint.Check() with invalid operator should return false")
	}
}

func TestVersion_Compare_AllEqual(t *testing.T) {
	// Test versions where all components are equal to hit compareInt return 0
	// for major, minor, and patch comparisons
	tests := []struct {
		name string
		a, b Version
		want int
	}{
		{
			name: "equal_major_minor_patch",
			a:    Version{Major: 1, Minor: 2, Patch: 3},
			b:    Version{Major: 1, Minor: 2, Patch: 3},
			want: 0,
		},
		{
			name: "equal_major_minor_different_patch",
			a:    Version{Major: 1, Minor: 2, Patch: 3},
			b:    Version{Major: 1, Minor: 2, Patch: 4},
			want: -1,
		},
		{
			name: "equal_major_different_minor",
			a:    Version{Major: 1, Minor: 2, Patch: 3},
			b:    Version{Major: 1, Minor: 3, Patch: 3},
			want: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Compare(tt.b); got != tt.want {
				t.Errorf("Compare() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestMatrix_ConcurrentAccess(t *testing.T) {
	// Test that Matrix is safe for concurrent access
	m := NewMatrix()

	// Pre-populate with some entries
	m.Add(Compatibility{
		Component:  "test",
		MinVersion: MustParse("1.0.0"),
	})

	const goroutines = 10
	const iterations = 100

	done := make(chan bool, goroutines*3)

	// Concurrent Add
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < iterations; j++ {
				m.Add(Compatibility{
					Component:  "component-" + string(rune('a'+id)),
					MinVersion: MustParse("1.0.0"),
				})
			}
			done <- true
		}(i)
	}

	// Concurrent Check
	for i := 0; i < goroutines; i++ {
		go func() {
			for j := 0; j < iterations; j++ {
				m.Check("test", MustParse("1.5.0"))
			}
			done <- true
		}()
	}

	// Concurrent Negotiate
	for i := 0; i < goroutines; i++ {
		go func() {
			available := []Version{
				MustParse("0.5.0"),
				MustParse("1.0.0"),
				MustParse("1.5.0"),
			}
			for j := 0; j < iterations; j++ {
				_, _ = m.Negotiate("test", available)
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < goroutines*3; i++ {
		<-done
	}
}
