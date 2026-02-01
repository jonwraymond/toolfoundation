package version_test

import (
	"fmt"
	"log"

	"github.com/jonwraymond/toolfoundation/version"
)

func ExampleParse() {
	v, err := version.Parse("1.2.3")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Major: %d, Minor: %d, Patch: %d\n", v.Major, v.Minor, v.Patch)

	// Also accepts 'v' prefix
	v2, _ := version.Parse("v2.0.0-beta.1")
	fmt.Printf("Version: %s, Prerelease: %s\n", v2.String(), v2.Prerelease)
	// Output:
	// Major: 1, Minor: 2, Patch: 3
	// Version: v2.0.0-beta.1, Prerelease: beta.1
}

func ExampleMustParse() {
	v := version.MustParse("1.0.0")
	fmt.Println(v.String())
	// Output:
	// v1.0.0
}

func ExampleVersion_Compare() {
	v1 := version.MustParse("1.0.0")
	v2 := version.MustParse("2.0.0")
	v3 := version.MustParse("1.0.0")

	fmt.Println("1.0.0 vs 2.0.0:", v1.Compare(v2))
	fmt.Println("2.0.0 vs 1.0.0:", v2.Compare(v1))
	fmt.Println("1.0.0 vs 1.0.0:", v1.Compare(v3))
	// Output:
	// 1.0.0 vs 2.0.0: -1
	// 2.0.0 vs 1.0.0: 1
	// 1.0.0 vs 1.0.0: 0
}

func ExampleVersion_LessThan() {
	v1 := version.MustParse("1.0.0")
	v2 := version.MustParse("2.0.0")

	if v1.LessThan(v2) {
		fmt.Println("Upgrade available!")
	}
	// Output:
	// Upgrade available!
}

func ExampleVersion_Compatible() {
	v1 := version.MustParse("1.5.0")
	v2 := version.MustParse("1.0.0")
	v3 := version.MustParse("2.0.0")

	// Same major version, v1 >= v2
	fmt.Println("1.5.0 compatible with 1.0.0:", v1.Compatible(v2))

	// Different major versions are incompatible
	fmt.Println("1.5.0 compatible with 2.0.0:", v1.Compatible(v3))
	// Output:
	// 1.5.0 compatible with 1.0.0: true
	// 1.5.0 compatible with 2.0.0: false
}

func ExampleParseConstraint() {
	constraints := []string{
		">=1.0.0",
		"<2.0.0",
		"^1.5.0",
		"~1.2.0",
	}

	testVersion := version.MustParse("1.8.0")

	for _, cs := range constraints {
		c, _ := version.ParseConstraint(cs)
		fmt.Printf("%s matches 1.8.0: %v\n", cs, c.Check(testVersion))
	}
	// Output:
	// >=1.0.0 matches 1.8.0: true
	// <2.0.0 matches 1.8.0: true
	// ^1.5.0 matches 1.8.0: true
	// ~1.2.0 matches 1.8.0: false
}

func ExampleConstraint_Check() {
	// Caret constraint: same major version
	caret, _ := version.ParseConstraint("^1.0.0")

	fmt.Println("^1.0.0 accepts 1.5.0:", caret.Check(version.MustParse("1.5.0")))
	fmt.Println("^1.0.0 accepts 2.0.0:", caret.Check(version.MustParse("2.0.0")))

	// Tilde constraint: same major.minor version
	tilde, _ := version.ParseConstraint("~1.2.0")

	fmt.Println("~1.2.0 accepts 1.2.5:", tilde.Check(version.MustParse("1.2.5")))
	fmt.Println("~1.2.0 accepts 1.3.0:", tilde.Check(version.MustParse("1.3.0")))
	// Output:
	// ^1.0.0 accepts 1.5.0: true
	// ^1.0.0 accepts 2.0.0: false
	// ~1.2.0 accepts 1.2.5: true
	// ~1.2.0 accepts 1.3.0: false
}

func ExampleMatrix() {
	matrix := version.NewMatrix()

	// Define compatibility requirements
	matrix.Add(version.Compatibility{
		Component:  "toolfoundation",
		MinVersion: version.MustParse("0.1.0"),
	})

	// Check a version
	ok, msg := matrix.Check("toolfoundation", version.MustParse("0.2.0"))
	fmt.Println("0.2.0 compatible:", ok)
	if msg == "" {
		fmt.Println("Message: (empty)")
	} else {
		fmt.Println("Message:", msg)
	}

	// Check an incompatible version
	ok, _ = matrix.Check("toolfoundation", version.MustParse("0.0.5"))
	fmt.Println("0.0.5 compatible:", ok)
	// Output:
	// 0.2.0 compatible: true
	// Message: (empty)
	// 0.0.5 compatible: false
}

func ExampleMatrix_Negotiate() {
	matrix := version.NewMatrix()
	matrix.Add(version.Compatibility{
		Component:  "api",
		MinVersion: version.MustParse("2.0.0"),
	})

	available := []version.Version{
		version.MustParse("1.0.0"),
		version.MustParse("2.0.0"),
		version.MustParse("2.5.0"),
		version.MustParse("3.0.0"),
	}

	best, err := matrix.Negotiate("api", available)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Best compatible version:", best.String())
	// Output:
	// Best compatible version: v3.0.0
}

func ExampleCompatibility_deprecated() {
	matrix := version.NewMatrix()

	// Mark a component as deprecated
	matrix.Add(version.Compatibility{
		Component:  "old-api",
		MinVersion: version.MustParse("1.0.0"),
		Deprecated: true,
		Message:    "Use new-api instead",
	})

	ok, msg := matrix.Check("old-api", version.MustParse("1.5.0"))
	fmt.Println("Still compatible:", ok)
	fmt.Println("Deprecation notice:", msg)
	// Output:
	// Still compatible: true
	// Deprecation notice: Use new-api instead
}
