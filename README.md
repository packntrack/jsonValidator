# JsonValidator

#### JsonValidator is a package that helps create CRUD APIs by validating HTTP JSON bodies against a predefined "form" struct.

### TL;DR
This is a basic example on how to use this package.
```go
type Object struct {
    ID             *string    `validations:"type=string;required=true`
    Code           *int       `validations:"type=int;choices=1,2,3`
    Person         *Person    `validations:"type=struct`
    Owners         []string   `validations:"type=[]string`
}

form := new(Object)
validationErrors := jsonValidator.Validate(c.Body(), form)
```


## Available validations

### Types

```go
type Object struct {
    Name           *string    `validations:"type=string`
    Code           *int       `validations:"type=int`
    Price          *float64   `validations:"type=float`
    Successful     *bool      `validations:"type=bool`
    Person         *Person    `validations:"type=struct`
    Owners         []string   `validations:"type=[]string`
    PreviousCodes  []int      `validations:"type=[]int`
    PreviousPrices []float64  `validations:"type=[]float`
    PersonList     []Person   `validations:"type=[]struct`
}
```
All types (besides the slices) need to be a pointer. This makes it clear what fields the user sent in the JSON.
As structs have zero-values, all the basic types would get the zero-value even though the user might not be sending any value.
The slices do not need this because the zero-value of it is nil.
The package is capable of transforming data if necessary.
For example if the form is ```type struct {Count int `validations:"type=int"`}``` and the received JSON is `{'count': '12345'}` the package will cast the '12345' string into an int.

### Required
```go
type Object struct {
    Name *string `validations:"type=string;required=true`
}
```

### Min and Max
```go
type Object struct {
    Name           *string    `validations:"type=string;min=1;max=10"`
    Code           *int       `validations:"type=int;min=1;max=10"`
    Price          *float64   `validations:"type=float;min=1;max=10"`
    Owners         []string   `validations:"type=[]string;min=1;max=2"`
    PreviousCodes  []int      `validations:"type=[]int;min=1;max=2"`
    PreviousPrices []float64  `validations:"type=[]float;min=1;max=2"`
    PersonList     []Person   `validations:"type=[]struct;min=1;max=2"`
}
```
Any type can have a min and max besides the 'bool' type.

The min and max functionality depends on each type:
- For `type=string` the min and max are the minimum/maximum length of the string.
- For `type=int` the min and max are the minimum/maximum number for the int.
- For `type=float` the min and max are the minimum/maximum number for the float.
- For `type=[]string` or `type=[]int` or `type=[]float` or `type=[]struct` the min and max are the minimum/maximum length for the array. Basically how many "options" can be selected.

### Choices
```go
type Object struct {
    Name           *string    `validations:"type=string;choices=Daniel,Jaime,Carolina"`
    Code           *int       `validations:"type=int;choices=1,2,3"`
    Price          *float64   `validations:"type=float;choices=1.0,2.0,3.0"`
    Owners         []string   `validations:"type=[]string;choices=Daniel,Jaime,Carolina"`
    PreviousCodes  []int      `validations:"type=[]int;choices=1,2,3"`
    PreviousPrices []float64  `validations:"type=[]float;choices=1.0,2.0,3.0"`
}
```
The package will validate the received JSON against the available choices

### Structs
```go
type Person struct {
    Name *string `validations:"type=string"`
    Age  *int    `validations:"type=int"`
}
type Object struct {
    Person     *Person   `validations:"type=struct"`
    PersonList []Person  `validations:"type=[]struct"`
}
```
This package is also capable of validating validations inside the defined struct


### Errors
Last but not least we have the errors. The package will return the errors in the ValidationError slice.
```go
type ValidationError struct {
	Field   string
	Message string
}
```