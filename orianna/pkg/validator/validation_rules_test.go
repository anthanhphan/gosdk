// Copyright (c) 2026 anthanhphan <an.thanhphan.work@gmail.com>

package validator

import (
	"reflect"
	"testing"
)

func TestValidateRequired(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		wantErr bool
	}{
		{
			name:    "non-empty string should not return error",
			field:   "Name",
			value:   reflect.ValueOf("John"),
			wantErr: false,
		},
		{
			name:    "empty string should return error",
			field:   "Name",
			value:   reflect.ValueOf(""),
			wantErr: true,
		},
		{
			name:    "non-zero int should not return error",
			field:   "Age",
			value:   reflect.ValueOf(25),
			wantErr: false,
		},
		{
			name:    "zero int should return error",
			field:   "Age",
			value:   reflect.ValueOf(0),
			wantErr: true,
		},
		{
			name:    "non-nil slice should not return error",
			field:   "Items",
			value:   reflect.ValueOf([]string{"item1"}),
			wantErr: false,
		},
		{
			name:    "nil slice should return error",
			field:   "Items",
			value:   reflect.ValueOf([]string(nil)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRequired(tt.field, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateRequired() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && err == nil {
				t.Error("validateRequired() should return error")
			}
		})
	}
}

func TestValidateMin(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		minVal  int
		wantErr bool
	}{
		{
			name:    "string with length >= min should not return error",
			field:   "Name",
			value:   reflect.ValueOf("John"),
			minVal:  3,
			wantErr: false,
		},
		{
			name:    "string with length < min should return error",
			field:   "Name",
			value:   reflect.ValueOf("Jo"),
			minVal:  3,
			wantErr: true,
		},
		{
			name:    "int value >= min should not return error",
			field:   "Age",
			value:   reflect.ValueOf(18),
			minVal:  18,
			wantErr: false,
		},
		{
			name:    "int value < min should return error",
			field:   "Age",
			value:   reflect.ValueOf(17),
			minVal:  18,
			wantErr: true,
		},
		{
			name:    "slice with length >= min should not return error",
			field:   "Items",
			value:   reflect.ValueOf([]string{"item1", "item2"}),
			minVal:  2,
			wantErr: false,
		},
		{
			name:    "slice with length < min should return error",
			field:   "Items",
			value:   reflect.ValueOf([]string{"item1"}),
			minVal:  2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMin(tt.field, tt.value, tt.minVal)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateMin() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMax(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		maxVal  int
		wantErr bool
	}{
		{
			name:    "string with length <= max should not return error",
			field:   "Name",
			value:   reflect.ValueOf("John"),
			maxVal:  10,
			wantErr: false,
		},
		{
			name:    "string with length > max should return error",
			field:   "Name",
			value:   reflect.ValueOf("Very Long Name"),
			maxVal:  10,
			wantErr: true,
		},
		{
			name:    "int value <= max should not return error",
			field:   "Age",
			value:   reflect.ValueOf(100),
			maxVal:  100,
			wantErr: false,
		},
		{
			name:    "int value > max should return error",
			field:   "Age",
			value:   reflect.ValueOf(101),
			maxVal:  100,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateMax(tt.field, tt.value, tt.maxVal)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateMax() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		wantErr bool
	}{
		{
			name:    "valid email should not return error",
			field:   "Email",
			value:   reflect.ValueOf("user@example.com"),
			wantErr: false,
		},
		{
			name:    "invalid email should return error",
			field:   "Email",
			value:   reflect.ValueOf("invalid-email"),
			wantErr: true,
		},
		{
			name:    "empty email should not return error (use required for that)",
			field:   "Email",
			value:   reflect.ValueOf(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEmail(tt.field, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateEmail() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateURL(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		wantErr bool
	}{
		{
			name:    "valid URL should not return error",
			field:   "URL",
			value:   reflect.ValueOf("https://example.com"),
			wantErr: false,
		},
		{
			name:    "invalid URL should return error",
			field:   "URL",
			value:   reflect.ValueOf("not-a-url"),
			wantErr: true,
		},
		{
			name:    "empty URL should not return error",
			field:   "URL",
			value:   reflect.ValueOf(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateURL(tt.field, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateURL() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateNumeric(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		wantErr bool
	}{
		{
			name:    "numeric string should not return error",
			field:   "Code",
			value:   reflect.ValueOf("12345"),
			wantErr: false,
		},
		{
			name:    "non-numeric string should return error",
			field:   "Code",
			value:   reflect.ValueOf("abc123"),
			wantErr: true,
		},
		{
			name:    "empty string should not return error",
			field:   "Code",
			value:   reflect.ValueOf(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateNumeric(tt.field, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateNumeric() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateAlpha(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		wantErr bool
	}{
		{
			name:    "alphabetic string should not return error",
			field:   "Name",
			value:   reflect.ValueOf("John"),
			wantErr: false,
		},
		{
			name:    "non-alphabetic string should return error",
			field:   "Name",
			value:   reflect.ValueOf("John123"),
			wantErr: true,
		},
		{
			name:    "empty string should not return error",
			field:   "Name",
			value:   reflect.ValueOf(""),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAlpha(tt.field, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("validateAlpha() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateOneOf(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		param   string
		wantErr bool
	}{
		{
			name:    "value in allowed list",
			field:   "Status",
			value:   reflect.ValueOf("active"),
			param:   "active inactive pending",
			wantErr: false,
		},
		{
			name:    "value not in allowed list",
			field:   "Status",
			value:   reflect.ValueOf("deleted"),
			param:   "active inactive pending",
			wantErr: true,
		},
		{
			name:    "empty value should not return error",
			field:   "Status",
			value:   reflect.ValueOf(""),
			param:   "active inactive",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateOneOf(tt.field, tt.value, tt.param)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateOneOf() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateLen(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		length  int
		wantErr bool
	}{
		{
			name:    "string with exact length",
			field:   "Code",
			value:   reflect.ValueOf("ABC"),
			length:  3,
			wantErr: false,
		},
		{
			name:    "string too short",
			field:   "Code",
			value:   reflect.ValueOf("AB"),
			length:  3,
			wantErr: true,
		},
		{
			name:    "string too long",
			field:   "Code",
			value:   reflect.ValueOf("ABCD"),
			length:  3,
			wantErr: true,
		},
		{
			name:    "slice with exact count",
			field:   "Items",
			value:   reflect.ValueOf([]string{"a", "b"}),
			length:  2,
			wantErr: false,
		},
		{
			name:    "slice with wrong count",
			field:   "Items",
			value:   reflect.ValueOf([]string{"a"}),
			length:  2,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateLen(tt.field, tt.value, tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateLen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateComparison(t *testing.T) {
	tests := []struct {
		name    string
		field   string
		value   reflect.Value
		param   string
		op      string
		wantErr bool
	}{
		// gt
		{name: "gt int pass", field: "Age", value: reflect.ValueOf(int64(10)), param: "5", op: "gt", wantErr: false},
		{name: "gt int fail", field: "Age", value: reflect.ValueOf(int64(5)), param: "5", op: "gt", wantErr: true},
		// gte
		{name: "gte int pass equal", field: "Age", value: reflect.ValueOf(int64(5)), param: "5", op: "gte", wantErr: false},
		{name: "gte int pass greater", field: "Age", value: reflect.ValueOf(int64(6)), param: "5", op: "gte", wantErr: false},
		{name: "gte int fail", field: "Age", value: reflect.ValueOf(int64(4)), param: "5", op: "gte", wantErr: true},
		// lt
		{name: "lt int pass", field: "Age", value: reflect.ValueOf(int64(4)), param: "5", op: "lt", wantErr: false},
		{name: "lt int fail", field: "Age", value: reflect.ValueOf(int64(5)), param: "5", op: "lt", wantErr: true},
		// lte
		{name: "lte int pass equal", field: "Age", value: reflect.ValueOf(int64(5)), param: "5", op: "lte", wantErr: false},
		{name: "lte int fail", field: "Age", value: reflect.ValueOf(int64(6)), param: "5", op: "lte", wantErr: true},
		// float
		{name: "gt float pass", field: "Score", value: reflect.ValueOf(3.14), param: "3.0", op: "gt", wantErr: false},
		{name: "lt float pass", field: "Score", value: reflect.ValueOf(2.5), param: "3.0", op: "lt", wantErr: false},
		{name: "lt float fail", field: "Score", value: reflect.ValueOf(3.5), param: "3.0", op: "lt", wantErr: true},
		// uint
		{name: "gt uint pass", field: "Count", value: reflect.ValueOf(uint64(10)), param: "5", op: "gt", wantErr: false},
		{name: "gt uint fail", field: "Count", value: reflect.ValueOf(uint64(3)), param: "5", op: "gt", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateComparison(tt.field, tt.value, tt.param, tt.op)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateComparison(%s) error = %v, wantErr %v", tt.op, err, tt.wantErr)
			}
		})
	}
}

// TestValidateIntegration tests the full Validate function with new rules
func TestValidateIntegration(t *testing.T) {
	type Order struct {
		Status   string  `validate:"required,oneof=pending active completed"`
		Quantity int     `validate:"required,gt=0,lte=1000"`
		Code     string  `validate:"len=6"`
		Price    float64 `validate:"gte=0"`
	}

	t.Run("valid order", func(t *testing.T) {
		order := Order{Status: "active", Quantity: 5, Code: "ABC123", Price: 9.99}
		if err := Validate(order); err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("invalid status", func(t *testing.T) {
		order := Order{Status: "cancelled", Quantity: 5, Code: "ABC123", Price: 9.99}
		err := Validate(order)
		if err == nil {
			t.Error("expected error for invalid status")
		}
	})

	t.Run("quantity zero", func(t *testing.T) {
		order := Order{Status: "active", Quantity: 0, Code: "ABC123", Price: 9.99}
		err := Validate(order)
		if err == nil {
			t.Error("expected error for zero quantity")
		}
	})

	t.Run("wrong code length", func(t *testing.T) {
		order := Order{Status: "active", Quantity: 5, Code: "AB", Price: 9.99}
		err := Validate(order)
		if err == nil {
			t.Error("expected error for wrong code length")
		}
	})

	t.Run("negative price", func(t *testing.T) {
		order := Order{Status: "active", Quantity: 5, Code: "ABC123", Price: -1.0}
		err := Validate(order)
		if err == nil {
			t.Error("expected error for negative price")
		}
	})
}
