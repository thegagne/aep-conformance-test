package checks

import (
	"regexp"
	"slices"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

// field is a discovered property with its dotted path from the resource root.
type field struct {
	Name   string
	Path   string
	Schema *discovery.Schema
}

// walkFields visits every property of a resource schema, recursing into nested
// object properties and array-of-object item properties. Each property is
// visited once with its dotted path.
func walkFields(s *discovery.Schema, prefix string, fn func(field)) {
	if s == nil || s.Properties == nil {
		return
	}
	for _, name := range s.Properties.Order {
		prop := s.Properties.Map[name]
		if prop == nil {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		fn(field{Name: name, Path: path, Schema: prop})
		switch prop.Type {
		case "object":
			walkFields(prop, path, fn)
		case "array":
			if prop.Items != nil && prop.Items.Type == "object" {
				walkFields(prop.Items, path, fn)
			}
		}
	}
}

var (
	snakeRE      = regexp.MustCompile(`^[a-z0-9_]+$`)
	badUnderRE   = regexp.MustCompile(`^_|_$|__`)
	wordDigitRE  = regexp.MustCompile(`(^|_)[0-9]`)
	reservedWord = map[string]bool{
		"new": true, "class": true, "function": true, "import": true,
		"return": true, "switch": true, "default": true, "for": true,
	}
)

// fieldCheck registers a MUST/SHOULD check that flags the first field failing
// pred. ok(f) returns false for a violating field.
func fieldCheck(id string, aep int, level Level, title, quoteText string, ok func(field) bool) {
	Register(Check{
		ID: id, AEP: aep, Level: level, Static: true, Title: title,
		Run: func(rc *RunContext) Result {
			var bad []string
			walkFields(rc.Resource.Schema, "", func(f field) {
				if !ok(f) {
					bad = append(bad, f.Path)
				}
			})
			if len(bad) == 0 {
				return pass("all fields conform")
			}
			return quote(failf("non-conforming field(s): %s", strings.Join(bad, ", ")), quoteText)
		},
	})
}

func init() {
	// AEP-140: field names are lower_snake_case.
	fieldCheck("field-lower-snake-case", 140, MUST,
		"Field names are lower_snake_case",
		"JSON and protobuf fields must use lower_snake_case names.",
		func(f field) bool { return snakeRE.MatchString(f.Name) })

	// AEP-140: no leading/trailing/adjacent underscores.
	fieldCheck("field-no-bad-underscores", 140, MUST,
		"Field names have no leading/trailing/adjacent underscores",
		"Fields must not contain leading, trailing, or adjacent underscores.",
		func(f field) bool { return !badUnderRE.MatchString(f.Name) })

	// AEP-140: each word must not begin with a number.
	fieldCheck("field-word-not-start-digit", 140, MUST,
		"No field-name word begins with a digit",
		"Each word in a field name must not begin with a number.",
		func(f field) bool { return !wordDigitRE.MatchString(f.Name) })

	// AEP-140: boolean fields should not carry an is_ prefix.
	fieldCheck("field-bool-no-is-prefix", 140, SHOULD,
		"Boolean fields omit the 'is_' prefix",
		"Boolean fields should omit the prefix 'is' (e.g. disabled, not is_disabled).",
		func(f field) bool { return !(f.Schema.Type == "boolean" && strings.HasPrefix(f.Name, "is_")) })

	// AEP-140: prefer uri over url.
	fieldCheck("field-uri-not-url", 140, SHOULD,
		"URL/URI fields use 'uri' not 'url'",
		"Field names representing URLs or URIs should always use uri rather than url.",
		func(f field) bool { return !hasWord(f.Name, "url") })

	// AEP-140: avoid programming-language keywords.
	fieldCheck("field-avoid-reserved-keywords", 140, SHOULD,
		"Field names avoid common reserved keywords",
		"Field names should avoid names likely to conflict with keywords in common programming languages.",
		func(f field) bool { return !reservedWord[f.Name] })

	// AEP-142: absolute-time fields should be named *_time.
	fieldCheck("field-timestamp-suffix-time", 142, SHOULD,
		"date-time fields are named with a _time suffix",
		"Fields representing time should have names ending in _time, such as create_time or update_time.",
		func(f field) bool {
			return !(f.Schema.Format == "date-time" && !strings.HasSuffix(f.Name, "_time"))
		})

	// AEP-142: time fields should not be past tense.
	fieldCheck("field-timestamp-not-past-tense", 142, SHOULD,
		"time fields are not past tense",
		"Time fields should not be named using the past tense (create_time, not created_time).",
		func(f field) bool {
			pastTense := f.Name == "created_time" || f.Name == "updated_time" ||
				f.Name == "deleted_time" || strings.HasSuffix(f.Name, "_at")
			return !pastTense
		})

	// AEP-141: item-count fields should use the _count suffix, not a num_ prefix.
	fieldCheck("field-count-suffix-not-num-prefix", 141, SHOULD,
		"Count fields use a _count suffix, not a num_ prefix",
		"If the quantity is a number of items, the field should use the suffix _count (not the prefix num_).",
		func(f field) bool {
			return !(strings.HasPrefix(f.Name, "num_") || strings.HasPrefix(f.Name, "number_of_"))
		})

	// AEP-148: create_time / update_time, when present, are output only.
	for _, tf := range []string{"create_time", "update_time"} {
		name := tf
		Register(Check{
			ID: "standard-field-" + name + "-output-only", AEP: 148, Level: SHOULD, Static: true,
			Title:      name + " is output only",
			Applicable: func(rc *RunContext) bool { return rc.Resource.Schema.HasProp(name) },
			Run: func(rc *RunContext) Result {
				if rc.Resource.Schema.Prop(name).ReadOnly {
					return pass(name + " is readOnly")
				}
				return quote(failf("%s is present but not readOnly", name),
					"Standard fields such as create_time and update_time are output only.")
			},
		})
	}
}

// hasWord reports whether w appears as a snake_case word within name.
func hasWord(name, w string) bool {
	return slices.Contains(strings.Split(name, "_"), w)
}
