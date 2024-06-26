package jsonValidator

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"testing"
)

type Errors []error

func (a Errors) Len() int           { return len(a) }
func (a Errors) Less(i, j int) bool { return a[i].Error() < a[j].Error() }
func (a Errors) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

func toStringPointer(s string) *string {
	return &s
}
func toIntPointer(i int) *int {
	return &i
}
func toFloatPointer(f float64) *float64 {
	return &f
}
func toBoolPointer(b bool) *bool {
	return &b
}

func TestTitleCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"test_1", "christopher george latore wallace", "Christopher George Latore Wallace"},
		{"test_2", "Christopher George Latore Wallace", "Christopher George Latore Wallace"},
		{"test_3", "christophergeorgelatorewallace", "Christophergeorgelatorewallace"},
		{"test_4", "christopherGeorgeLatoreWallace", "ChristopherGeorgeLatoreWallace"},
		{"test_5", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TitleCase(tt.input); got != tt.want {
				t.Errorf("TitleCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLowerCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"test_1", "Christopher George Latore Wallace", "christopher george latore wallace"},
		{"test_2", "christopher george latore wallace", "christopher george latore wallace"},
		{"test_3", "christophergeorgelatorewallace", "christophergeorgelatorewallace"},
		{"test_4", "ChristopherGeorgeLatoreWallace", "christopherGeorgeLatoreWallace"},
		{"test_5", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LowerCase(tt.input); got != tt.want {
				t.Errorf("LowerCase() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidate_BasicTypes(t *testing.T) {
	type createObject struct {
		Name            *string   `validations:"type=string"`
		Code            *int      `validations:"type=int"`
		Price           *float64  `validations:"type=float"`
		Successful      *bool     `validations:"type=bool"`
		Owners          []string  `validations:"type=[]string"`
		PreviousCodes   []int     `validations:"type=[]int"`
		PreviousPrices  []float64 `validations:"type=[]float"`
		PreviousPrices2 []float64 `validations:"type=[]float"`
	}
	type input struct {
		jsonData []byte
		form     *createObject
	}
	type want struct {
		errors []error
		form   createObject
	}
	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "test_all_types",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\", \"code\": 123, \"price\": 12.3, \"successful\": true, \"owners\": [\"Daniel\", \"Silva\"], \"previousCodes\": [1, 2, 3], \"previousPrices\": [1.1, 2.2, 3.3]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("Daniel"),
					Code:           toIntPointer(123),
					Price:          toFloatPointer(12.3),
					Successful:     toBoolPointer(true),
					Owners:         []string{"Daniel", "Silva"},
					PreviousCodes:  []int{1, 2, 3},
					PreviousPrices: []float64{1.1, 2.2, 3.3},
				},
			},
		},
		{
			name: "test_string_types",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\", \"code\": \"123\", \"price\": \"12.3\", \"successful\": \"true\", \"owners\": [\"Daniel\", \"Silva\"], \"previousCodes\": [\"1\", \"2\", \"3\"], \"previousPrices\": [\"1.1\", \"2.2\", \"3.3\"]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("Daniel"),
					Code:           toIntPointer(123),
					Price:          toFloatPointer(12.3),
					Successful:     toBoolPointer(true),
					Owners:         []string{"Daniel", "Silva"},
					PreviousCodes:  []int{1, 2, 3},
					PreviousPrices: []float64{1.1, 2.2, 3.3},
				},
			},
		},
		{
			name: "test_int_types",
			input: input{
				jsonData: []byte("{\"name\": 123, \"code\": 123, \"price\": 123, \"successful\": 1, \"owners\": [123, 456], \"previousCodes\": [1, 2, 3], \"previousPrices\": [1, 2, 3]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("123"),
					Code:           toIntPointer(123),
					Price:          toFloatPointer(123),
					Successful:     toBoolPointer(true),
					Owners:         []string{"123", "456"},
					PreviousCodes:  []int{1, 2, 3},
					PreviousPrices: []float64{1, 2, 3},
				},
			},
		},
		{
			name: "test_float_types",
			input: input{
				jsonData: []byte("{\"name\": 12.3, \"code\": 12.0, \"price\": 12.3, \"successful\": 1.0, \"owners\": [12.3, 45.6], \"previousCodes\": [1.0, 2.0, 3.0], \"previousPrices\": [1.1, 2.2, 3.3]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("12.3"),
					Code:           toIntPointer(12),
					Price:          toFloatPointer(12.3),
					Successful:     toBoolPointer(true),
					Owners:         []string{"12.3", "45.6"},
					PreviousCodes:  []int{1, 2, 3},
					PreviousPrices: []float64{1.1, 2.2, 3.3},
				},
			},
		},
		{
			name: "test_invalid_json",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\",}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{ValidationError{Field: "json", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], "{\"name\": \"Daniel\",}")}},
				form:   createObject{},
			},
		},
		{
			name: "test_invalid_field_name",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\", \"surname\": \"Silva\"}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{ValidationError{
					Field:   "surname",
					Message: DefaultMessages["InvalidField"],
				}},
				form: createObject{Name: toStringPointer("Daniel")},
			},
		},
		{
			name: "test_types_error",
			input: input{
				jsonData: []byte("{\"name\": [], \"code\": \"Daniel\", \"price\": \"Daniel\", \"successful\": 123, \"owners\": [[]], \"previousCodes\": [\"Daniel\"], \"previousPrices\": [\"Daniel\"], \"previousPrices2\": {}}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], []any{})},
					ValidationError{Field: "code", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], "Daniel")},
					ValidationError{Field: "price", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], "Daniel")},
					ValidationError{Field: "successful", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], 123)},
					ValidationError{Field: "owners[0]", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], []any{})},
					ValidationError{Field: "previousCodes[0]", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], "Daniel")},
					ValidationError{Field: "previousPrices[0]", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], "Daniel")},
					ValidationError{Field: "previousPrices2", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], map[string]string{})},
				},
				form: createObject{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.input.jsonData, tt.input.form)

			// Sort
			sort.Sort(Errors(got))
			sort.Sort(Errors(tt.want.errors))

			inputForm, _ := json.Marshal(*tt.input.form)
			wantForm, _ := json.Marshal(tt.want.form)

			if !reflect.DeepEqual(got, tt.want.errors) {
				t.Errorf("Validate() = %v, want %v", got, tt.want.errors)
			}
			if !reflect.DeepEqual(*tt.input.form, tt.want.form) {
				t.Errorf("Validate() = %v, want %v", string(inputForm), string(wantForm))
			}
		})
	}
}

func TestValidate_Required(t *testing.T) {
	type Person struct {
		Name *string `validations:"type=string;required=true"`
		Age  *int    `validations:"type=int"`
	}
	type createObject struct {
		Name           *string   `validations:"type=string;required=true"`
		Code           *int      `validations:"type=int;required=true"`
		Price          *float64  `validations:"type=float;required=true"`
		Successful     *bool     `validations:"type=bool;required=true"`
		Person         *Person   `validations:"type=struct;required=true"`
		Owners         []string  `validations:"type=[]string;required=true"`
		PreviousCodes  []int     `validations:"type=[]int;required=true"`
		PreviousPrices []float64 `validations:"type=[]float;required=true"`
		PersonList     []Person  `validations:"type=[]struct;required=true"`
	}
	type input struct {
		jsonData []byte
		form     *createObject
	}
	type want struct {
		errors []error
		form   createObject
	}
	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "test_required",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\", \"code\": 123, \"price\": 12.3, \"successful\": true, \"person\": {\"name\": \"Daniel\"}, \"owners\": [\"Daniel\", \"Silva\"], \"previousCodes\": [1, 2, 3], \"previousPrices\": [1.1, 2.2, 3.3], \"personList\": [{\"name\": \"Silva\"}]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("Daniel"),
					Code:           toIntPointer(123),
					Price:          toFloatPointer(12.3),
					Successful:     toBoolPointer(true),
					Person:         &Person{Name: toStringPointer("Daniel")},
					Owners:         []string{"Daniel", "Silva"},
					PreviousCodes:  []int{1, 2, 3},
					PreviousPrices: []float64{1.1, 2.2, 3.3},
					PersonList:     []Person{{Name: toStringPointer("Silva")}},
				},
			},
		},
		{
			name: "test_required_errors",
			input: input{
				jsonData: []byte("{}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "code", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "price", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "successful", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "person", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "owners", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "previousCodes", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "previousPrices", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "personList", Message: DefaultMessages["RequiredField"]},
				},
				form: createObject{},
			},
		},
		{
			name: "test_required_struct_errors",
			input: input{
				jsonData: []byte("{\"person\": {\"age\": 26}, \"personList\": [{\"age\": 26}]}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "code", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "price", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "successful", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "person.name", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "owners", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "previousCodes", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "previousPrices", Message: DefaultMessages["RequiredField"]},
					ValidationError{Field: "personList[0].name", Message: DefaultMessages["RequiredField"]},
				},
				form: createObject{Person: &Person{Age: toIntPointer(26)}, PersonList: []Person{}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.input.jsonData, tt.input.form)

			// Sort
			sort.Sort(Errors(got))
			sort.Sort(Errors(tt.want.errors))

			inputForm, _ := json.Marshal(*tt.input.form)
			wantForm, _ := json.Marshal(tt.want.form)

			if !reflect.DeepEqual(got, tt.want.errors) {
				t.Errorf("Validate() = %v, want %v", got, tt.want.errors)
			}
			if !reflect.DeepEqual(*tt.input.form, tt.want.form) {
				t.Errorf("Validate() = %v, want %v", string(inputForm), string(wantForm))
			}
		})
	}
}

func TestValidate_MinMax(t *testing.T) {
	type Person struct {
		Name *string `validations:"type=string;required=true"`
		Age  *int    `validations:"type=int;required=true"`
	}
	type createObject struct {
		Name           *string   `validations:"type=string;min=1;max=10"`
		Code           *int      `validations:"type=int;min=1;max=10"`
		Price          *float64  `validations:"type=float;min=1;max=10"`
		Owners         []string  `validations:"type=[]string;min=1;max=2"`
		PreviousCodes  []int     `validations:"type=[]int;min=1;max=2"`
		PreviousPrices []float64 `validations:"type=[]float;min=1;max=2"`
		PersonList     []Person  `validations:"type=[]struct;min=1;max=2"`
	}
	type input struct {
		jsonData []byte
		form     *createObject
	}
	type want struct {
		errors []error
		form   createObject
	}
	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "test_min",
			input: input{
				jsonData: []byte("{\"name\": \"D\", \"code\": 1, \"price\": 1.0, \"owners\": [\"Daniel\"], \"previousCodes\": [1], \"previousPrices\": [1.0], \"personList\": [{\"name\": \"Jose\", \"age\": 20}]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("D"),
					Code:           toIntPointer(1),
					Price:          toFloatPointer(1.0),
					Owners:         []string{"Daniel"},
					PreviousCodes:  []int{1},
					PreviousPrices: []float64{1.0},
					PersonList:     []Person{{Name: toStringPointer("Jose"), Age: toIntPointer(20)}},
				},
			},
		},
		{
			name: "test_max",
			input: input{
				jsonData: []byte("{\"name\": \"JoseDaniel\", \"code\": 10, \"price\": 10.0, \"owners\": [\"Daniel\", \"Silva\"], \"previousCodes\": [10, 9], \"previousPrices\": [10.0, 9.0], \"personList\": [{\"name\": \"Jose\", \"age\": 20}, {\"name\": \"Silva\", \"age\": 32}]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("JoseDaniel"),
					Code:           toIntPointer(10),
					Price:          toFloatPointer(10.0),
					Owners:         []string{"Daniel", "Silva"},
					PreviousCodes:  []int{10, 9},
					PreviousPrices: []float64{10.0, 9.0},
					PersonList:     []Person{{Name: toStringPointer("Jose"), Age: toIntPointer(20)}, {Name: toStringPointer("Silva"), Age: toIntPointer(32)}},
				},
			},
		},
		{
			name: "test_min_error",
			input: input{
				jsonData: []byte("{\"name\": \"\", \"code\": 0, \"price\": 0.0, \"owners\": [], \"previousCodes\": [], \"previousPrices\": [], \"personList\": []}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: fmt.Sprintf(DefaultMessages["InvalidMinString"], 1)},
					ValidationError{Field: "code", Message: fmt.Sprintf(DefaultMessages["InvalidMinNumber"], 1)},
					ValidationError{Field: "price", Message: fmt.Sprintf(DefaultMessages["InvalidMinNumber"], 1)},
					ValidationError{Field: "owners", Message: fmt.Sprintf(DefaultMessages["InvalidMinList"], 1)},
					ValidationError{Field: "previousCodes", Message: fmt.Sprintf(DefaultMessages["InvalidMinList"], 1)},
					ValidationError{Field: "previousPrices", Message: fmt.Sprintf(DefaultMessages["InvalidMinList"], 1)},
					ValidationError{Field: "personList", Message: fmt.Sprintf(DefaultMessages["InvalidMinList"], 1)},
				},
				form: createObject{},
			},
		},
		{
			name: "test_max_error",
			input: input{
				jsonData: []byte("{\"name\": \"JoseDanielSilva\", \"code\": 11, \"price\": 11.0, \"owners\": [\"1\", \"2\", \"3\"], \"previousCodes\": [1, 2, 3], \"previousPrices\": [1.0, 2.0, 3.0], \"personList\": [{\"name\": \"Jose\", \"age\": 20}, {\"name\": \"Daniel\", \"age\": 26}, {\"name\": \"Silva\", \"age\": 32}]}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: fmt.Sprintf(DefaultMessages["InvalidMaxString"], 10)},
					ValidationError{Field: "code", Message: fmt.Sprintf(DefaultMessages["InvalidMaxNumber"], 10)},
					ValidationError{Field: "price", Message: fmt.Sprintf(DefaultMessages["InvalidMaxNumber"], 10)},
					ValidationError{Field: "owners", Message: fmt.Sprintf(DefaultMessages["InvalidMaxList"], 2)},
					ValidationError{Field: "previousCodes", Message: fmt.Sprintf(DefaultMessages["InvalidMaxList"], 2)},
					ValidationError{Field: "previousPrices", Message: fmt.Sprintf(DefaultMessages["InvalidMaxList"], 2)},
					ValidationError{Field: "personList", Message: fmt.Sprintf(DefaultMessages["InvalidMaxList"], 2)},
				},
				form: createObject{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.input.jsonData, tt.input.form)

			// Sort
			sort.Sort(Errors(got))
			sort.Sort(Errors(tt.want.errors))

			inputForm, _ := json.Marshal(*tt.input.form)
			wantForm, _ := json.Marshal(tt.want.form)

			if !reflect.DeepEqual(got, tt.want.errors) {
				t.Errorf("Validate() = %v, want %v", got, tt.want.errors)
			}
			if !reflect.DeepEqual(*tt.input.form, tt.want.form) {
				t.Errorf("Validate() = %v, want %v", string(inputForm), string(wantForm))
			}
		})
	}
}

func TestValidate_Choices(t *testing.T) {
	type createObject struct {
		Name           *string   `validations:"type=string;choices=Daniel"`
		Code           *int      `validations:"type=int;choices=1,2"`
		Price          *float64  `validations:"type=float;choices=1.0,2.0"`
		Owners         []string  `validations:"type=[]string;choices=Daniel"`
		PreviousCodes  []int     `validations:"type=[]int;choices=1,2"`
		PreviousPrices []float64 `validations:"type=[]float;choices=1.0,2.0"`
	}
	type input struct {
		jsonData []byte
		form     *createObject
	}
	type want struct {
		errors []error
		form   createObject
	}
	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "test_choices_1",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\", \"code\": 1, \"price\": 1.0, \"owners\": [\"Daniel\"], \"previousCodes\": [1, 2], \"previousPrices\": [1.0, 2.0]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("Daniel"),
					Code:           toIntPointer(1),
					Price:          toFloatPointer(1.0),
					Owners:         []string{"Daniel"},
					PreviousCodes:  []int{1, 2},
					PreviousPrices: []float64{1.0, 2.0},
				},
			},
		},
		{
			name: "test_choices_2",
			input: input{
				jsonData: []byte("{\"name\": \"Daniel\", \"code\": 2, \"price\": 2.0, \"owners\": [\"Daniel\"], \"previousCodes\": [2], \"previousPrices\": [2.0]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Name:           toStringPointer("Daniel"),
					Code:           toIntPointer(2),
					Price:          toFloatPointer(2.0),
					Owners:         []string{"Daniel"},
					PreviousCodes:  []int{2},
					PreviousPrices: []float64{2.0},
				},
			},
		},
		{
			name: "test_choices_errors_1",
			input: input{
				jsonData: []byte("{\"name\": \"Daniele\", \"code\": 101, \"price\": 101.0, \"owners\": [\"Jose\", \"Magalhaes\"], \"previousCodes\": [3, 4], \"previousPrices\": [3.0, 4.0]}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], "Daniele", []string{"Daniel"})},
					ValidationError{Field: "code", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 101, []string{"1", "2"})},
					ValidationError{Field: "price", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 101.0, []string{"1", "2"})},
					ValidationError{Field: "owners[0]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], "Jose", []string{"Daniel"})},
					ValidationError{Field: "owners[1]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], "Magalhaes", []string{"Daniel"})},
					ValidationError{Field: "previousCodes[0]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 3, []string{"1", "2"})},
					ValidationError{Field: "previousCodes[1]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 4, []string{"1", "2"})},
					ValidationError{Field: "previousPrices[0]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 3.0, []string{"1", "2"})},
					ValidationError{Field: "previousPrices[1]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 4.0, []string{"1", "2"})},
				},
				form: createObject{},
			},
		},
		{
			name: "test_choices_errors_2",
			input: input{
				jsonData: []byte("{\"name\": \"Jose\", \"code\": 10, \"price\": 10.0, \"owners\": [\"Jose\", \"Silva\"], \"previousCodes\": [1, 3], \"previousPrices\": [1.0, 3.0]}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "name", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], "Jose", []string{"Daniel"})},
					ValidationError{Field: "code", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 10, []string{"1", "2"})},
					ValidationError{Field: "price", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 10.0, []string{"1", "2"})},
					ValidationError{Field: "owners[0]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], "Jose", []string{"Daniel"})},
					ValidationError{Field: "owners[1]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], "Silva", []string{"Daniel"})},
					ValidationError{Field: "previousCodes[1]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 3, []string{"1", "2"})},
					ValidationError{Field: "previousPrices[1]", Message: fmt.Sprintf(DefaultMessages["InvalidChoice"], 3.0, []string{"1", "2"})},
				},
				form: createObject{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.input.jsonData, tt.input.form)

			// Sort
			sort.Sort(Errors(got))
			sort.Sort(Errors(tt.want.errors))

			inputForm, _ := json.Marshal(*tt.input.form)
			wantForm, _ := json.Marshal(tt.want.form)

			if !reflect.DeepEqual(got, tt.want.errors) {
				t.Errorf("Validate() = %v, want %v", got, tt.want.errors)
			}
			if !reflect.DeepEqual(*tt.input.form, tt.want.form) {
				t.Errorf("Validate() = %v, want %v", string(inputForm), string(wantForm))
			}
		})
	}
}

func TestValidate_Struct(t *testing.T) {
	type Person struct {
		Name *string `validations:"type=string"`
		Age  *int    `validations:"type=int"`
	}
	type createObject struct {
		Person      *Person  `validations:"type=struct"`
		PersonList  []Person `validations:"type=[]struct"`
		PersonList2 []Person `validations:"type=[]struct"`
	}
	type input struct {
		jsonData []byte
		form     *createObject
	}
	type want struct {
		errors []error
		form   createObject
	}
	tests := []struct {
		name  string
		input input
		want  want
	}{
		{
			name: "test_structs",
			input: input{
				jsonData: []byte("{\"person\": {\"name\": \"Daniel\", \"age\": 26}, \"personList\": [{\"name\": \"Jose\", \"age\": 20}, {\"name\": \"Silva\", \"age\": 32}]}"),
				form:     new(createObject),
			},
			want: want{
				errors: nil,
				form: createObject{
					Person: &Person{Name: toStringPointer("Daniel"), Age: toIntPointer(26)},
					PersonList: []Person{
						{Name: toStringPointer("Jose"), Age: toIntPointer(20)},
						{Name: toStringPointer("Silva"), Age: toIntPointer(32)},
					},
				},
			},
		},
		{
			name: "test_structs_errors",
			input: input{
				jsonData: []byte("{\"person\": {\"firstName\": \"Daniel\", \"age\": 26}, \"personList\": [{\"firstName\": \"Jose\", \"age\": 20}], \"personList2\": {}}"),
				form:     new(createObject),
			},
			want: want{
				errors: []error{
					ValidationError{Field: "person.firstName", Message: DefaultMessages["InvalidField"]},
					ValidationError{Field: "personList[0].firstName", Message: DefaultMessages["InvalidField"]},
					ValidationError{Field: "personList2", Message: fmt.Sprintf(DefaultMessages["InvalidFormat"], map[string]string{})},
				},
				form: createObject{
					Person:     &Person{Age: toIntPointer(26)},
					PersonList: []Person{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Validate(tt.input.jsonData, tt.input.form)

			// Sort
			sort.Sort(Errors(got))
			sort.Sort(Errors(tt.want.errors))

			if !reflect.DeepEqual(got, tt.want.errors) {
				t.Errorf("Validate() = %v, want %v", got, tt.want.errors)
			}
			if !reflect.DeepEqual(*tt.input.form, tt.want.form) {
				t.Errorf("Validate() = %v, want %v", *tt.input.form, tt.want.form)
			}
		})
	}
}

func TestValidationError_Error(t *testing.T) {
	tests := []struct {
		name            string
		validationError ValidationError
		want            string
	}{
		{
			name:            "test_1",
			validationError: ValidationError{Field: "test_field", Message: "test message."},
			want:            "Field test_field: test message.",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vr := ValidationError{
				Field:   tt.validationError.Field,
				Message: tt.validationError.Message,
			}
			if got := vr.Error(); got != tt.want {
				t.Errorf("Error() = %v, want %v", got, tt.want)
			}
		})
	}
}
