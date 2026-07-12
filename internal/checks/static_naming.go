package checks

import (
	"regexp"
	"strings"

	"github.com/thegagne/aep-conformance-test/internal/discovery"
)

var apiNameRE = regexp.MustCompile(`^[a-z][a-z0-9.\-/]*$`)

func init() {
	// AEP-102: API name is lowercase, starts with a letter, valid characters.
	Register(Check{
		ID: "api-name-format", AEP: 102, Level: MUST, Static: true, PerAPI: true,
		Title:      "API name is lowercase and uses valid characters",
		Applicable: func(rc *RunContext) bool { return rc.Model.APIName != "" },
		Run: func(rc *RunContext) Result {
			n := rc.Model.APIName
			if apiNameRE.MatchString(n) {
				return pass("api name: " + n)
			}
			return quote(failf("api name %q is not lowercase RFC-1035 (^[a-z][a-z0-9.\\-/]*$)", n),
				"The API Name must use all lowercase, start with a lowercase letter, and only use valid domain name characters.")
		},
	})

	// AEP-4: resource type is {apiName}/{TypeName}.
	Register(Check{
		ID: "resource-type-format", AEP: 4, Level: MUST, Static: true,
		Title:      "Resource type is {api-name}/{type}",
		Applicable: func(rc *RunContext) bool { return rc.Resource.Type != "" },
		Run: func(rc *RunContext) Result {
			t := rc.Resource.Type
			if i := strings.LastIndex(t, "/"); i > 0 && i < len(t)-1 {
				return pass("type: " + t)
			}
			return quote(failf("resource type %q is not in {api-name}/{type} form", t),
				"Resource types must be of the form {API Name}/{Type Name}.")
		},
	})

	// AEP-130: operationId format for standard non-list methods.
	stdMethods := map[discovery.Method]string{
		discovery.MethodGet:    "Get",
		discovery.MethodCreate: "Create",
		discovery.MethodUpdate: "Update",
		discovery.MethodApply:  "Apply",
		discovery.MethodDelete: "Delete",
	}
	Register(Check{
		ID: "operation-id-standard", AEP: 130, Level: MUST, Static: true,
		Title: "Standard-method operationIds are {Method}{Singular}",
		Run: func(rc *RunContext) Result {
			for m, prefix := range stdMethods {
				ep := rc.Resource.Method(m)
				if ep == nil || ep.OperationID == "" {
					continue
				}
				want := prefix + pascal(rc.Resource.Singular)
				if ep.OperationID != want {
					return quote(failf("%s operationId is %q, want %q", m, ep.OperationID, want),
						"The operationId must follow {standard-method-name}{resource-singular} for non-list standard methods.")
				}
			}
			return pass("standard operationIds well-formed")
		},
	})

	// AEP-130: List operationId is List{Plural}.
	Register(Check{
		ID: "operation-id-list-plural", AEP: 130, Level: MUST, Static: true,
		Title: "List operationId is List{Plural}",
		Applicable: func(rc *RunContext) bool {
			ep := rc.Resource.Method(discovery.MethodList)
			return ep != nil && ep.OperationID != ""
		},
		Run: func(rc *RunContext) Result {
			ep := rc.Resource.Method(discovery.MethodList)
			want := "List" + pascal(rc.Resource.Plural)
			if ep.OperationID == want {
				return pass(ep.OperationID)
			}
			return quote(failf("List operationId is %q, want %q", ep.OperationID, want),
				"The operationId must follow List{resource-plural} for lists.")
		},
	})
}

// pascal converts a kebab/snake/space identifier to PascalCase.
func pascal(s string) string {
	fields := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	var b strings.Builder
	for _, f := range fields {
		if f == "" {
			continue
		}
		b.WriteString(strings.ToUpper(f[:1]))
		b.WriteString(f[1:])
	}
	return b.String()
}
