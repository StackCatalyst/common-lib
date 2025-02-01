package testing

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

// AssertionResult represents the result of an assertion evaluation
type AssertionResult struct {
	// Assertion is the original assertion string
	Assertion string
	// Success indicates if the assertion passed
	Success bool
	// Message contains details about the assertion result
	Message string
}

// AssertionContext contains data needed for assertion evaluation
type AssertionContext struct {
	// Variables contains the current variable values
	Variables map[string]interface{}
	// Outputs contains the current output values
	Outputs map[string]interface{}
	// Resources contains the created resources
	Resources []*Resource
}

// EvaluateAssertion evaluates a single assertion
func EvaluateAssertion(assertion string, ctx *AssertionContext) *AssertionResult {
	result := &AssertionResult{
		Assertion: assertion,
	}

	// Parse assertion components
	parts := strings.Fields(assertion)
	if len(parts) < 3 {
		result.Success = false
		result.Message = "invalid assertion format: must contain at least 3 parts"
		return result
	}

	// Get value based on reference type
	var actualValue interface{}
	var expectedValue string
	switch parts[0] {
	case "variable":
		actualValue = ctx.Variables[parts[1]]
		if len(parts) > 3 && parts[2] != "exists" {
			expectedValue = parts[3]
		}
	case "output":
		actualValue = ctx.Outputs[parts[1]]
		if len(parts) > 3 {
			expectedValue = parts[3]
		}
	case "resource":
		if len(parts) < 4 {
			result.Success = false
			result.Message = "invalid resource assertion format: must contain at least 4 parts"
			return result
		}
		actualValue = findResourceProperty(ctx.Resources, parts[1], parts[2])
		if len(parts) > 4 {
			expectedValue = parts[4]
		}
		parts = []string{parts[0], parts[1], parts[3]} // Adjust parts for condition evaluation
		if expectedValue != "" {
			parts = append(parts, expectedValue)
		}
	default:
		result.Success = false
		result.Message = fmt.Sprintf("unknown reference type: %s", parts[0])
		return result
	}

	// Evaluate condition
	switch parts[2] {
	case "equals", "=":
		if expectedValue == "" {
			result.Success = false
			result.Message = "missing expected value for equals condition"
			return result
		}
		result.Success = evaluateEquals(actualValue, expectedValue)
		result.Message = fmt.Sprintf("expected %v to equal %v", actualValue, expectedValue)

	case "contains":
		if expectedValue == "" {
			result.Success = false
			result.Message = "missing expected value for contains condition"
			return result
		}
		result.Success = evaluateContains(actualValue, expectedValue)
		result.Message = fmt.Sprintf("expected %v to contain %v", actualValue, expectedValue)

	case "matches":
		if expectedValue == "" {
			result.Success = false
			result.Message = "missing expected value for matches condition"
			return result
		}
		result.Success = evaluateMatches(actualValue, expectedValue)
		result.Message = fmt.Sprintf("expected %v to match pattern %v", actualValue, expectedValue)

	case "exists":
		result.Success = actualValue != nil
		result.Message = fmt.Sprintf("expected %s to exist", parts[1])

	case "type":
		if expectedValue == "" {
			result.Success = false
			result.Message = "missing expected value for type condition"
			return result
		}
		result.Success = evaluateType(actualValue, expectedValue)
		result.Message = fmt.Sprintf("expected %v to be of type %s", actualValue, expectedValue)

	default:
		result.Success = false
		result.Message = fmt.Sprintf("unknown condition: %s", parts[2])
	}

	return result
}

// findResourceProperty finds a property value in a list of resources
func findResourceProperty(resources []*Resource, id, property string) interface{} {
	for _, r := range resources {
		if r.ID == id {
			if strings.Contains(property, ".") {
				// Handle nested properties
				parts := strings.Split(property, ".")
				var value interface{} = r.Properties
				for _, part := range parts {
					if m, ok := value.(map[string]interface{}); ok {
						value = m[part]
						if value == nil {
							return nil
						}
					} else {
						return nil
					}
				}
				return value
			}
			return r.Properties[property]
		}
	}
	return nil
}

// evaluateEquals compares two values for equality
func evaluateEquals(actual interface{}, expected string) bool {
	// Handle different types
	switch v := actual.(type) {
	case string:
		return v == expected
	case int:
		if n, err := strconv.Atoi(expected); err == nil {
			return v == n
		}
	case float64:
		if n, err := strconv.ParseFloat(expected, 64); err == nil {
			return v == n
		}
	case bool:
		if b, err := strconv.ParseBool(expected); err == nil {
			return v == b
		}
	}
	return false
}

// evaluateContains checks if a value contains another value
func evaluateContains(actual interface{}, expected string) bool {
	switch v := actual.(type) {
	case string:
		return strings.Contains(v, expected)
	case []interface{}:
		for _, item := range v {
			if fmt.Sprintf("%v", item) == expected {
				return true
			}
		}
	case map[string]interface{}:
		_, exists := v[expected]
		return exists
	}
	return false
}

// evaluateMatches checks if a value matches a pattern
func evaluateMatches(actual interface{}, pattern string) bool {
	str := fmt.Sprintf("%v", actual)
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	return re.MatchString(str)
}

// evaluateType checks if a value is of the expected type
func evaluateType(actual interface{}, expectedType string) bool {
	if actual == nil {
		return false
	}
	actualType := reflect.TypeOf(actual).String()
	return strings.EqualFold(actualType, expectedType)
}

// CompareValues compares two values with type conversion
func CompareValues(actual, expected interface{}) (bool, string) {
	if actual == nil && expected == nil {
		return true, ""
	}
	if actual == nil || expected == nil {
		return false, fmt.Sprintf("expected %v but got %v", expected, actual)
	}

	// Convert expected value to actual type if possible
	actualValue := reflect.ValueOf(actual)
	expectedValue := reflect.ValueOf(expected)

	switch actualValue.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if expectedValue.Kind() == reflect.String {
			if n, err := strconv.ParseInt(expectedValue.String(), 10, 64); err == nil {
				return actualValue.Int() == n, ""
			}
		}
	case reflect.Float32, reflect.Float64:
		if expectedValue.Kind() == reflect.String {
			if n, err := strconv.ParseFloat(expectedValue.String(), 64); err == nil {
				return actualValue.Float() == n, ""
			}
		}
	case reflect.Bool:
		if expectedValue.Kind() == reflect.String {
			if b, err := strconv.ParseBool(expectedValue.String()); err == nil {
				return actualValue.Bool() == b, ""
			}
		}
	case reflect.String:
		return actualValue.String() == fmt.Sprintf("%v", expected), ""
	case reflect.Slice, reflect.Array:
		if expectedValue.Kind() == reflect.Slice || expectedValue.Kind() == reflect.Array {
			if actualValue.Len() != expectedValue.Len() {
				return false, "length mismatch"
			}
			for i := 0; i < actualValue.Len(); i++ {
				if eq, _ := CompareValues(actualValue.Index(i).Interface(), expectedValue.Index(i).Interface()); !eq {
					return false, fmt.Sprintf("mismatch at index %d", i)
				}
			}
			return true, ""
		}
	case reflect.Map:
		if expectedValue.Kind() == reflect.Map {
			if actualValue.Len() != expectedValue.Len() {
				return false, "length mismatch"
			}
			for _, key := range expectedValue.MapKeys() {
				actualMapValue := actualValue.MapIndex(key)
				if !actualMapValue.IsValid() {
					return false, fmt.Sprintf("missing key %v", key)
				}
				if eq, _ := CompareValues(actualMapValue.Interface(), expectedValue.MapIndex(key).Interface()); !eq {
					return false, fmt.Sprintf("mismatch at key %v", key)
				}
			}
			return true, ""
		}
	}

	// If types are different and no conversion was possible
	if actualValue.Type() != expectedValue.Type() {
		return false, fmt.Sprintf("type mismatch: expected %T but got %T", expected, actual)
	}

	// Direct comparison for same types
	return reflect.DeepEqual(actual, expected), ""
}
