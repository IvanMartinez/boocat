package formats

//@TODO: Merge formats and validators?

import (
	"context"

	"github.com/ivanmartinez/boocat/validators"
)

// InitializeFields initializes the format fields
func InitializeFields() {
	Formats = make(map[string]Format)

	Formats["author"] = Format{
		Name: "author",
		Fields: map[string]validators.Validate{
			"name":      nil,
			"birthdate": nil,
			"biography": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "biography": {}},
	}

	Formats["book"] = Format{
		Name: "book",
		Fields: map[string]validators.Validate{
			"name":     nil,
			"year":     nil,
			"author":   nil,
			"synopsis": nil,
		},
		Searchable: map[string]struct{}{"name": {}, "synopsis": {}},
	}
}

func InitializeValidators() {
	Formats["author"].Fields["name"] = ValidateAnything
	Formats["author"].Fields["birthdate"] = ValidateAnything
	Formats["author"].Fields["biography"] = ValidateAnything

	Formats["book"].Fields["name"] = ValidateAnything
	Formats["book"].Fields["year"] = ValidateAnything
	Formats["book"].Fields["author"] = ValidateAnything
	Formats["book"].Fields["synopsis"] = ValidateAnything
}

func ValidateAnything(_ context.Context, _ interface{}) bool {
	return true
}
