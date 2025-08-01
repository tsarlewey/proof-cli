package utils

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Helper functions for the CLI application
// Add your utility functions here

// FormatOutput formats the output based on the given format
func FormatOutput(data any, format string) (string, error) {
	// Implement formatting logic based on your needs
	return "", nil
}

// BuildQueryParams converts a struct to URL query parameters using reflection
func BuildQueryParams(params any) url.Values {
	queryParams := url.Values{}

	if params == nil {
		return queryParams
	}

	v := reflect.ValueOf(params)

	// Handle pointer to struct
	if v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return queryParams
		}
		v = v.Elem()
	}

	// Only work with structs
	if v.Kind() != reflect.Struct {
		return queryParams
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		fieldType := t.Field(i)

		// Skip unexported fields
		if !field.CanInterface() {
			continue
		}

		// Get the JSON tag for the parameter name
		tag := fieldType.Tag.Get("json")
		if tag == "" || tag == "-" {
			continue
		}

		// Remove omitempty and other options from tag
		paramName := tag
		if idx := strings.Index(tag, ","); idx != -1 {
			paramName = tag[:idx]
		}

		// Skip if parameter name is empty
		if paramName == "" {
			continue
		}

		// Handle different field types
		switch field.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if value := field.Int(); value > 0 {
				queryParams.Set(paramName, strconv.FormatInt(value, 10))
			}
		case reflect.String:
			if value := field.String(); value != "" {
				queryParams.Set(paramName, value)
			}
		case reflect.Ptr:
			if !field.IsNil() {
				switch field.Elem().Kind() {
				case reflect.Struct:
					// Handle time.Time pointers
					if timeVal, ok := field.Interface().(*time.Time); ok && timeVal != nil {
						queryParams.Set(paramName, timeVal.Format(time.RFC3339))
					}
				}
			}
		}
	}

	return queryParams
}
