package jsonschema

import (
	"fmt"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canonical"
)

func (s *Schema) validateArray(arr []any, path string, errs *[]Error) {
	s.validateArraySize(arr, path, errs)
	s.validateUniqueItems(arr, path, errs)
	s.validateContains(arr, path, errs)
	if s.items != nil {
		for i, e := range arr {
			s.items.validate(e, joinPath(path, fmt.Sprintf("%d", i)), errs)
		}
	}
}

func (s *Schema) validateArraySize(arr []any, path string, errs *[]Error) {
	if s.hasMinItems && len(arr) < s.minItems {
		*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("array shorter than minItems %d", s.minItems)})
	}
	if s.hasMaxItems && len(arr) > s.maxItems {
		*errs = append(*errs, Error{Path: path, Msg: fmt.Sprintf("array longer than maxItems %d", s.maxItems)})
	}
}

func (s *Schema) validateUniqueItems(arr []any, path string, errs *[]Error) {
	if s.uniqueItems {
		seen := map[string]bool{}
		for _, e := range arr {
			key, err := canonical.MarshalString(e)
			if err == nil && seen[key] {
				*errs = append(*errs, Error{Path: path, Msg: "array items must be unique"})
				break
			}
			seen[key] = true
		}
	}
}

func (s *Schema) validateContains(arr []any, path string, errs *[]Error) {
	if s.contains != nil {
		for _, e := range arr {
			if len(s.contains.Validate(e)) == 0 {
				return
			}
		}
		*errs = append(*errs, Error{Path: path, Msg: "no array item matches contains"})
	}
}
