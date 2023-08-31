package jsonValidator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

func getValidations(formValue reflect.Value) map[string]*Validations {

	// 1) Initialize validations map and required fields map
	validationsMap := make(map[string]*Validations)

	// 3) Iterate over the form value.
	for i := 0; i < formValue.NumField(); i++ {

		// 3.1) Get field from form value.
		field := formValue.Type().Field(i)

		// 3.2) Get the validation using the tag "validations".
		validationsTag := field.Tag.Get(DefaultTagName)

		// 3.3) Split the validations in the tag by ";".
		validationsSplit := strings.Split(validationsTag, DefaultSeparator)

		// 3.4) Parse validations tags
		validations := parseValidationTags(validationsSplit)

		// 3.5) Update validations map with the validations from this field
		validationsMap[LowerCase(field.Name)] = validations
	}

	// 4) Return validations instance
	return validationsMap

}

func parseValidationTags(validationsSplit []string) *Validations {

	// 1) Initialize the validation instance.
	validations := new(Validations)

	// 2) Iterate over the validationSplit list to update the validations instance.
	for _, validation := range validationsSplit {

		// 2.1) Case: Required.
		if value, exists := strings.CutPrefix(validation, "required="); exists {
			if value == "true" {
				validations.Required = true
			}
		}

		// 2.2) Case: Type.
		if value, exists := strings.CutPrefix(validation, "type="); exists {
			switch value {
			case "string", "int", "float", "bool", "struct", "[]string", "[]int", "[]float", "[]struct":
				validations.Type = value
			}
		}

		// 2.3) Case: Min.
		if value, exists := strings.CutPrefix(validation, "min="); exists {
			switch validations.Type {
			case "string", "int", "[]string", "[]int", "[]float", "[]struct":
				if min, err := strconv.ParseInt(value, 10, 0); err == nil {
					validations.Min = float64(min)
				}
			case "float":
				if min, err := strconv.ParseFloat(value, 0); err == nil {
					validations.Min = min
				}
			}
		}

		// 2.4) Case: Max.
		if value, exists := strings.CutPrefix(validation, "max="); exists {
			switch validations.Type {
			case "string", "int", "[]string", "[]int", "[]float", "[]struct":
				if min, err := strconv.ParseInt(value, 10, 0); err == nil {
					validations.Max = float64(min)
				}
			case "float":
				if min, err := strconv.ParseFloat(value, 0); err == nil {
					validations.Max = min
				}
			}
		}

		// 2.5) Case: Choices.
		if value, exists := strings.CutPrefix(validation, "choices="); exists {
			if value != "" {
				var choices []any
				for _, choice := range strings.Split(value, DefaultChoicesSeparator) {
					switch validations.Type {
					case "string", "[]string":
						choices = append(choices, choice)
					case "int", "[]int":
						if intChoice, err := strconv.ParseInt(choice, 10, 0); err == nil {
							choices = append(choices, int(intChoice))
						}
					case "float", "[]float":
						if floatChoice, err := strconv.ParseFloat(choice, 0); err == nil {
							choices = append(choices, floatChoice)
						}
					}
				}
				validations.Choices = choices
			}
		}
	}

	// 3) Return the validations.
	return validations
}

func parseField(validations *Validations, fieldName string, fieldValue any, form reflect.Value) []error {
	switch validations.Type {
	case "string":
		return validateString(validations, fieldName, fieldValue, form)
	case "int":
		return validateInt(validations, fieldName, fieldValue, form)
	case "float":
		return validateFloat(validations, fieldName, fieldValue, form)
	case "bool":
		return validateBool(fieldName, fieldValue, form)
	case "struct":
		return validateStruct(fieldName, fieldValue, form)
	case "[]string":
		return validateList[string](validations, fieldName, fieldValue, form, validateStringType)
	case "[]int":
		return validateList[int](validations, fieldName, fieldValue, form, validateIntType)
	case "[]float":
		return validateList[float64](validations, fieldName, fieldValue, form, validateFloatType)
	default:
		return nil
	}
}

func validateString(validations *Validations, fieldName string, fieldValue any, form reflect.Value) []error {

	// 1) Initialize the errors list.
	var errors []error

	// 2) Validate fieldValue type.
	value, invalidFormat := validateStringType(fieldValue)
	if invalidFormat {
		errors = append(errors, ValidationError{Field: fieldName, Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], fieldValue)})
		return errors
	}

	// 3) Validate min and max.
	if !reflect.ValueOf(validations.Min).IsZero() && len(value) < int(validations.Min) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMinString"], int(validations.Min)),
		})
	}
	if !reflect.ValueOf(validations.Max).IsZero() && len(value) > int(validations.Max) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMaxString"], int(validations.Max)),
		})
	}

	// 4) Validate choices.
	if !reflect.ValueOf(validations.Choices).IsZero() && !contains[string](validations.Choices, value) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], value, validations.Choices),
		})
	}
	if errors != nil {
		return errors
	}

	// 5) Update form with the received value.
	form.FieldByName(TitleCase(fieldName)).Set(reflect.ValueOf(value))

	// 6) Return errors.
	return errors
}

func validateStringType(fieldValue any) (string, bool) {

	// 1) Initialize variables.
	var invalidFormat = true
	var value string

	// 2) Validate fieldValue type.
	switch v := fieldValue.(type) {
	case string:
		value = v
		invalidFormat = false
	case float64, int, bool:
		value = fmt.Sprintf("%v", v)
		invalidFormat = false
	}

	// 3) Return.
	return value, invalidFormat
}

func validateInt(validations *Validations, fieldName string, fieldValue any, form reflect.Value) []error {

	// 1) Initialize the errors list.
	var errors []error

	// 2) Validate the fieldValue type.
	value, invalidFormat := validateIntType(fieldValue)
	if invalidFormat {
		errors = append(errors, ValidationError{Field: fieldName, Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], fieldValue)})
		return errors
	}

	// 3) Validate min and max.
	if !reflect.ValueOf(validations.Min).IsZero() && value < int(validations.Min) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMinNumber"], int(validations.Min)),
		})
	}
	if !reflect.ValueOf(validations.Max).IsZero() && value > int(validations.Max) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMaxNumber"], int(validations.Max)),
		})
	}

	// 4) Validate choices.
	if !reflect.ValueOf(validations.Choices).IsZero() && !contains[int](validations.Choices, value) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], value, validations.Choices),
		})
	}
	if errors != nil {
		return errors
	}

	// 5) Update form with the received value.
	form.FieldByName(TitleCase(fieldName)).Set(reflect.ValueOf(value))

	// 6) Return errors.
	return errors
}

func validateIntType(fieldValue any) (int, bool) {

	// 1) Initialize variables.
	var invalidFormat = true
	var value int

	// 2) Validate fieldValue type.
	switch v := fieldValue.(type) {
	case string:
		intValue, err := strconv.ParseInt(v, 10, 0)
		if err == nil {
			invalidFormat = false
			value = int(intValue)
		}
	case float64:
		castedValue := int(v)
		if float64(castedValue) == v {
			value = castedValue
			invalidFormat = false
		}
	case int:
		value = v
		invalidFormat = false
	}

	// 3) Return.
	return value, invalidFormat
}

func validateFloat(validations *Validations, fieldName string, fieldValue any, form reflect.Value) []error {

	// 1) Initialize the errors list.
	var errors []error

	// 2) Validate the fieldValue type.
	value, invalidFormat := validateFloatType(fieldValue)
	if invalidFormat {
		errors = append(errors, ValidationError{Field: fieldName, Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], fieldValue)})
		return errors
	}

	// 3) Validate min and max.
	if !reflect.ValueOf(validations.Min).IsZero() && value < validations.Min {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMinNumber"], validations.Min),
		})
	}
	if !reflect.ValueOf(validations.Max).IsZero() && value > validations.Max {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMaxNumber"], validations.Max),
		})
	}

	// 4) Validate choices.
	if !reflect.ValueOf(validations.Choices).IsZero() && !contains[float64](validations.Choices, value) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], value, validations.Choices),
		})
	}
	if errors != nil {
		return errors
	}

	// 5) Update form with the received value.
	form.FieldByName(TitleCase(fieldName)).Set(reflect.ValueOf(value))

	// 6) Return errors.
	return errors
}

func validateFloatType(fieldValue any) (float64, bool) {

	// 1) Initialize variables.
	var invalidFormat = true
	var value float64

	// 2) Validate fieldValue type.
	switch v := fieldValue.(type) {
	case string:
		valueParsed, err := strconv.ParseFloat(v, 0)
		if err == nil {
			value = valueParsed
			invalidFormat = false
		}
	case float64:
		value = v
		invalidFormat = false
	case int:
		value = float64(v)
		invalidFormat = false
	}

	// 3) Return.
	return value, invalidFormat
}

func validateBool(fieldName string, fieldValue any, form reflect.Value) []error {

	// 1) Initialize the errors list.
	var errors []error

	// 2) Validate the fieldValue type.
	value, invalidFormat := validateBoolType(fieldValue)
	if invalidFormat {
		errors = append(errors, ValidationError{Field: fieldName, Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], fieldValue)})
		return errors
	}

	// 3) Update form with the received value.
	form.FieldByName(TitleCase(fieldName)).Set(reflect.ValueOf(value))

	// 4) Return errors.
	return nil
}

func validateBoolType(fieldValue any) (bool, bool) {

	// 1) Initialize variables.
	var invalidFormat = true
	var value bool

	// 2) Validate fieldValue type.
	switch v := fieldValue.(type) {
	case string, int, float64:
		parsed := fmt.Sprintf("%v", v)
		boolValue, err := strconv.ParseBool(parsed)
		if err == nil {
			value = boolValue
			invalidFormat = false
		}
	case bool:
		value = v
		invalidFormat = false
	}

	// 3) Return.
	return value, invalidFormat
}

func validateStruct(fieldName string, fieldValue any, form reflect.Value) []error {

	// 1) Get field from the form and the inner struct.
	field := form.FieldByName(TitleCase(fieldName))

	// 2) Parse the value to []byte.
	jsonData, _ := json.Marshal(fieldValue)

	// 3) Get validations map.
	validationsMap := getValidations(field)

	errors := validateJsonData(jsonData, field, validationsMap)

	// 4) Return errors.
	return errors
}

func validateList[T string | int | float64](validations *Validations, fieldName string, fieldValue any, form reflect.Value, validateElement func(any) (T, bool)) []error {

	// 1) Initialize an errors list.
	var errors []error

	// 2) Validate fieldValue type.
	value, ok := fieldValue.([]any)
	if !ok {
		return append(errors, ValidationError{Field: fieldName, Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], fieldValue)})
	}

	// 3) Validate min and max.
	if !reflect.ValueOf(validations.Min).IsZero() && len(value) < int(validations.Min) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMinList"], int(validations.Min)),
		})
	}
	if !reflect.ValueOf(validations.Max).IsZero() && len(value) > int(validations.Max) {
		errors = append(errors, ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf(DefaultMessages["InvalidMaxList"], int(validations.Max)),
		})
	}
	if errors != nil {
		return errors
	}

	// 4) Parse elements.
	parsedValues, errors := parseElements[T](fieldName, value, validateElement)
	if errors != nil {
		return errors
	}

	// 5) Remove duplicate.
	parsedValues = removeDuplicate[T](parsedValues)

	// 6) Validate choices.
	errors = validateListChoices[T](fieldName, validations.Choices, parsedValues)
	if errors != nil {
		return errors
	}

	// 7) Update the form with the parsed values.
	form.FieldByName(TitleCase(fieldName)).Set(reflect.ValueOf(parsedValues))

	// 8) Return errors.
	return nil
}

func parseElements[T string | int | float64](fieldName string, valuesList []any, validateElement func(any) (T, bool)) ([]T, []error) {

	// 1) Initialize errors list and values parsed list.
	var errors []error
	var parsedValues []T

	// 2) Iterate over the values list received.
	for _, element := range valuesList {

		// 2.1) Validate the element.
		elemValue, invalidFormat := validateElement(element)

		// 2.2) If the element has an invalid format, add the error to the errors list.
		if invalidFormat {
			errors = append(errors, ValidationError{Field: fieldName, Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], element)})
		}

		// 2.3) Add the value to the values parsed list.
		parsedValues = append(parsedValues, elemValue)
	}

	// 3) Return the parsed values and the errors.
	return parsedValues, errors
}

func validateListChoices[T string | int | float64](fieldName string, choices []any, parsedValues []T) []error {

	// 1) Initialize an errors list.
	var errors []error

	// 2) If we have received choices, validate them.
	if !reflect.ValueOf(choices).IsZero() {
		for _, element := range parsedValues {
			if !contains[T](choices, element) {
				errors = append(errors, ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], element, choices),
				})
			}
		}
	}

	// 3) Return the errors.
	return errors
}

func removeDuplicate[T string | int | float64](sliceList []T) []T {
	allKeys := make(map[T]bool)
	var list []T
	for _, item := range sliceList {
		if _, value := allKeys[item]; !value {
			allKeys[item] = true
			list = append(list, item)
		}
	}
	return list
}

func contains[T string | int | float64](sliceList []any, value T) bool {
	for _, element := range sliceList {
		if value == element.(T) {
			return true
		}
	}
	return false
}
