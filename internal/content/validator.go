package content

import (
	"encoding/json"
	"errors"
	"fmt"
	"cmsapp/internal/db"
)

type Field struct {
	Name     string  `json:"name"`
	Type     string  `json:"type"`
	Required bool    `json:"required"`

	MinLength *int  `json:"minLength,omitempty"`
	MaxLength *int  `json:"maxLength,omitempty"`
	Min       *int  `json:"min,omitempty"`
	Max       *int  `json:"max,omitempty"`
	Unique    bool  `json:"unique,omitempty"`
	Fields []Field `json:"fields,omitempty"`

	TargetModel string `json:"targetModel,omitempty"`
}

type Schema struct {
	Fields []Field `json:"fields"`
}
func validateField(field Field, val interface{}) error {
	switch field.Type {
	case "string":
		strVal, ok := val.(string)
		if !ok {
			return fmt.Errorf("field %s must be string", field.Name)
		}
		if field.MinLength != nil && len(strVal) < *field.MinLength {
			return fmt.Errorf("field %s must have at least %d characters", field.Name, *field.MinLength)
		}
		if field.MaxLength != nil && len(strVal) > *field.MaxLength {
			return fmt.Errorf("field %s must have at most %d characters", field.Name, *field.MaxLength)
		}
		if field.Unique {
			var count int64
			db.DB.Model(&ContentEntry{}).
				Where(fmt.Sprintf("data ->> '%s' = ?", field.Name), strVal).
				Count(&count)
			if count > 0 {
				return fmt.Errorf("field %s must be unique", field.Name)
			}
		}

	case "number":
		numVal, ok := val.(float64)
		if !ok {
			return fmt.Errorf("field %s must be number", field.Name)
		}
		if field.Min != nil && int(numVal) < *field.Min {
			return fmt.Errorf("field %s must be >= %d", field.Name, *field.Min)
		}
		if field.Max != nil && int(numVal) > *field.Max {
			return fmt.Errorf("field %s must be <= %d", field.Name, *field.Max)
		}

	case "date":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("field %s must be date string", field.Name)
		}

	case "boolean":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("field %s must be boolean", field.Name)
		}

	case "relation":
		switch v := val.(type) {
		case float64:
			if v <= 0 {
				return fmt.Errorf("field %s must be a valid ID", field.Name)
			}
		case []interface{}:
			for _, id := range v {
				if idFloat, ok := id.(float64); !ok || idFloat <= 0 {
					return fmt.Errorf("field %s has invalid relation ID", field.Name)
				}
			}
		default:
			return fmt.Errorf("field %s must be number or array of numbers", field.Name)
		}

	case "object":
		obj, ok := val.(map[string]interface{})
		if !ok {
			return fmt.Errorf("field %s must be object", field.Name)
		}
		for _, subField := range field.Fields {
			subVal, has := obj[subField.Name]
			if subField.Required && !has {
				return fmt.Errorf("missing required field: %s.%s", field.Name, subField.Name)
			}
			if has {
				if err := validateField(subField, subVal); err != nil {
					return fmt.Errorf("%s.%s: %s", field.Name, subField.Name, err.Error())
				}
			}
		}

	default:
		return fmt.Errorf("unsupported field type: %s", field.Type)
	}

	return nil
}

func ValidateEntry(schemaData []byte, entryData []byte) error {
	var schema Schema
	if err := json.Unmarshal(schemaData, &schema); err != nil {
		return err
	}

	var entry map[string]interface{}
	if err := json.Unmarshal(entryData, &entry); err != nil {
		return err
	}

	schemaFields := make(map[string]Field)
	for _, f := range schema.Fields {
		schemaFields[f.Name] = f
	}

	for _, field := range schema.Fields {
		val, ok := entry[field.Name]

		if field.Required && !ok {
			return fmt.Errorf("missing required field: %s", field.Name)
		}

		if ok {
			if err := validateField(field, val); err != nil {
				return err
			}
		}
	}

	for key := range entry {
		if _, ok := schemaFields[key]; !ok {
			return errors.New("field " + key + " is not allowed by schema")
		}
	}

	return nil
}
