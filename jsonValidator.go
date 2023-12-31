package jsonValidator

import (
	"encoding/json"
	"fmt"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	"reflect"
	"strings"
	"unicode"
)

type ValidationError struct {
	Field   string
	Message string
}

func (vr ValidationError) Error() string {
	return fmt.Sprintf("Field %s: %s", vr.Field, vr.Message)
}

type Validations struct {
	Type     string
	Required bool
	Min      float64
	Max      float64
	Choices  []any
}

var DefaultMessages = map[string]string{
	"InvalidField":     "This field is invalid.",
	"InvalidFormat":    "This field has an invalid format (%v).",
	"InvalidMinString": "This field must have at least %v characters.",
	"InvalidMaxString": "This field must not have more than %v characters.",
	"InvalidMinNumber": "This field must be bigger than %v.",
	"InvalidMaxNumber": "This field must be smaller than %v.",
	"InvalidMinList":   "This field must have at least %v elements.",
	"InvalidMaxList":   "This field must not have more than %v elements.",
	"RequiredField":    "This field is required.",
	"InvalidChoice":    "This field has an invalid choice (%v). The valid choices are (%v)",
}

var DefaultTagName = "validations"
var DefaultSeparator = ";"
var DefaultChoicesSeparator = ","

func TitleCase(str string) string {
	return cases.Title(language.English, cases.NoLower).String(str)
}

func LowerCase(str string) string {

	if str == "" {
		return str
	}

	var result []string
	for _, s := range strings.Split(str, " ") {
		a := []rune(s)
		a[0] = unicode.ToLower(a[0])
		s = string(a)
		result = append(result, s)
	}
	return strings.Join(result, " ")
}

// Validate validates the json data against a form received and update the form with the parsed data.
func Validate(jsonData []byte, form any) []error {

	// 1) Get form value.
	formValue := reflect.ValueOf(form).Elem()

	// 2) Get all the validations from the form.
	validationsMap := getValidations(formValue)

	// 3) Validate JSON data.
	errors := validateJsonData(jsonData, formValue, validationsMap, "")

	// 4) Return the errors.
	return errors
}

func validateJsonData(jsonData []byte, form reflect.Value, validationsMap map[string]*Validations, parent string) []error {

	// 1) Initialize errors list.
	var errors []error

	// 2) Decode the json data into a decodedJson map.
	var decodedJson map[string]any
	err := json.Unmarshal(jsonData, &decodedJson)
	if err != nil {
		errors = append(errors, ValidationError{
			Field:   "json",
			Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], string(jsonData)),
		})
		return errors
	}

	// 3) Iterate over each key in the decodeJson map.
	for fieldName, fieldValue := range decodedJson {

		// 3.1) Get the validations for the given fieldName.
		validations, ok := validationsMap[fieldName]
		if !ok {
			errors = append(errors, ValidationError{
				Field:   getFieldName(parent, fieldName),
				Message: DefaultMessages["InvalidField"],
			})
			continue
		}

		// 3.2) Update the required bool to false since we have the field present.
		validations.Required = false

		// 3.3) Parse and validate the field against the defined validations.
		if validationsErrors := parseField(validations, fieldName, fieldValue, form, parent); validationsErrors != nil {
			errors = append(errors, validationsErrors...)
		}
	}

	// 4) Check if all the required fields were sent.
	for fieldName, validations := range validationsMap {
		if validations.Required {
			errors = append(errors, ValidationError{
				Field:   getFieldName(parent, fieldName),
				Message: DefaultMessages["RequiredField"],
			})
		}
	}

	// 5) Return the errors.
	return errors
}
