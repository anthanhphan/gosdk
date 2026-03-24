package validator

import (
	"testing"
)

// ============================================================================
// Core Rules
// ============================================================================

func TestRequired(t *testing.T) {
	type S struct {
		Str   string            `validate:"required"`
		Num   int               `validate:"required"`
		U     uint              `validate:"required"`
		F     float64           `validate:"required"`
		B     bool              `validate:"required"`
		Slice []int             `validate:"required"`
		Map   map[string]string `validate:"required"`
		Ptr   *int              `validate:"required"`
	}

	n := 1
	valid := S{Str: "a", Num: 1, U: 1, F: 1.0, B: true, Slice: []int{1}, Map: map[string]string{"k": "v"}, Ptr: &n}
	if err := Validate(valid); err != nil {
		t.Errorf("valid should pass: %v", err)
	}

	// All zero values should fail
	err := Validate(S{})
	if err == nil {
		t.Fatal("empty should fail")
	}
	errs := err.(ValidationErrors)
	if len(errs) < 7 {
		t.Errorf("expected at least 7 errors, got %d", len(errs))
	}
}

func TestMin_String(t *testing.T) {
	type S struct {
		Name string `validate:"min=3"`
	}
	if err := Validate(S{Name: "abc"}); err != nil {
		t.Errorf("min=3 pass: %v", err)
	}
	if err := Validate(S{Name: "ab"}); err == nil {
		t.Error("min=3 should fail for 2 chars")
	}
	// Empty string also fails min=3 (len 0 < 3)
	if err := Validate(S{Name: ""}); err == nil {
		t.Error("min=3 should fail for empty string")
	}
}

func TestMin_Slice(t *testing.T) {
	type S struct {
		Items []int `validate:"min=2"`
	}
	if err := Validate(S{Items: []int{1, 2}}); err != nil {
		t.Errorf("min slice pass: %v", err)
	}
	if err := Validate(S{Items: []int{1}}); err == nil {
		t.Error("min slice should fail")
	}
}

func TestMin_Int(t *testing.T) {
	type S struct {
		Age int `validate:"min=18"`
	}
	if err := Validate(S{Age: 18}); err != nil {
		t.Errorf("min int pass: %v", err)
	}
	if err := Validate(S{Age: 17}); err == nil {
		t.Error("min int should fail")
	}
}

func TestMin_Uint(t *testing.T) {
	type S struct {
		Count uint `validate:"min=5"`
	}
	if err := Validate(S{Count: 5}); err != nil {
		t.Errorf("min uint pass: %v", err)
	}
	if err := Validate(S{Count: 4}); err == nil {
		t.Error("min uint should fail")
	}
}

func TestMin_Float(t *testing.T) {
	type S struct {
		Score float64 `validate:"min=10"`
	}
	if err := Validate(S{Score: 10.0}); err != nil {
		t.Errorf("min float pass: %v", err)
	}
	if err := Validate(S{Score: 9.9}); err == nil {
		t.Error("min float should fail")
	}
}

func TestMax_String(t *testing.T) {
	type S struct {
		Name string `validate:"max=5"`
	}
	if err := Validate(S{Name: "abcde"}); err != nil {
		t.Errorf("max pass: %v", err)
	}
	if err := Validate(S{Name: "abcdef"}); err == nil {
		t.Error("max should fail")
	}
}

func TestMax_Slice(t *testing.T) {
	type S struct {
		Items []int `validate:"max=2"`
	}
	if err := Validate(S{Items: []int{1, 2}}); err != nil {
		t.Errorf("max slice pass: %v", err)
	}
	if err := Validate(S{Items: []int{1, 2, 3}}); err == nil {
		t.Error("max slice should fail")
	}
}

func TestMax_Int(t *testing.T) {
	type S struct {
		Age int `validate:"max=120"`
	}
	if err := Validate(S{Age: 120}); err != nil {
		t.Errorf("max int pass: %v", err)
	}
	if err := Validate(S{Age: 121}); err == nil {
		t.Error("max int should fail")
	}
}

func TestMax_Uint(t *testing.T) {
	type S struct {
		Count uint `validate:"max=10"`
	}
	if err := Validate(S{Count: 10}); err != nil {
		t.Errorf("max uint pass: %v", err)
	}
	if err := Validate(S{Count: 11}); err == nil {
		t.Error("max uint should fail")
	}
}

func TestMax_Uint_NegativeMax(t *testing.T) {
	// maxVal < 0 should always fail for uint (since uint is always >= 0)
	type S struct {
		Count uint `validate:"max=-1"`
	}
	// intFactory will fail to parse negative as Atoi("-1") = -1, which is valid
	// maxVal < 0 branch should trigger
	if err := Validate(S{Count: 0}); err == nil {
		t.Error("uint with max=-1 should fail")
	}
}

func TestMax_Float(t *testing.T) {
	type S struct {
		Score float64 `validate:"max=100"`
	}
	if err := Validate(S{Score: 100.0}); err != nil {
		t.Errorf("max float pass: %v", err)
	}
	if err := Validate(S{Score: 100.1}); err == nil {
		t.Error("max float should fail")
	}
}

func TestLen_String(t *testing.T) {
	type S struct {
		Code string `validate:"len=4"`
	}
	if err := Validate(S{Code: "ABCD"}); err != nil {
		t.Errorf("len pass: %v", err)
	}
	if err := Validate(S{Code: "ABC"}); err == nil {
		t.Error("len should fail for 3")
	}
	if err := Validate(S{Code: "ABCDE"}); err == nil {
		t.Error("len should fail for 5")
	}
}

func TestLen_Slice(t *testing.T) {
	type S struct {
		Items []int `validate:"len=3"`
	}
	if err := Validate(S{Items: []int{1, 2, 3}}); err != nil {
		t.Errorf("len slice pass: %v", err)
	}
	if err := Validate(S{Items: []int{1, 2}}); err == nil {
		t.Error("len slice should fail")
	}
}

func TestLen_Map(t *testing.T) {
	type S struct {
		M map[string]int `validate:"len=2"`
	}
	if err := Validate(S{M: map[string]int{"a": 1, "b": 2}}); err != nil {
		t.Errorf("len map pass: %v", err)
	}
	if err := Validate(S{M: map[string]int{"a": 1}}); err == nil {
		t.Error("len map should fail")
	}
}

func TestOneOf_String(t *testing.T) {
	type S struct {
		Status string `validate:"oneof=active inactive"`
	}
	if err := Validate(S{Status: "active"}); err != nil {
		t.Errorf("oneof pass: %v", err)
	}
	if err := Validate(S{Status: "deleted"}); err == nil {
		t.Error("oneof should fail")
	}
	// Empty string should pass
	if err := Validate(S{Status: ""}); err != nil {
		t.Errorf("oneof empty pass: %v", err)
	}
}

func TestOneOf_Int(t *testing.T) {
	type S struct {
		Code int `validate:"oneof=1 2 3"`
	}
	if err := Validate(S{Code: 2}); err != nil {
		t.Errorf("oneof int pass: %v", err)
	}
	if err := Validate(S{Code: 5}); err == nil {
		t.Error("oneof int should fail")
	}
}

func TestOneOf_Uint(t *testing.T) {
	type S struct {
		Level uint `validate:"oneof=1 2 3"`
	}
	if err := Validate(S{Level: 1}); err != nil {
		t.Errorf("oneof uint pass: %v", err)
	}
	if err := Validate(S{Level: 9}); err == nil {
		t.Error("oneof uint should fail")
	}
}

func TestOneOf_Float(t *testing.T) {
	type S struct {
		Rate float64 `validate:"oneof=1.5 2.5"`
	}
	if err := Validate(S{Rate: 1.5}); err != nil {
		t.Errorf("oneof float pass: %v", err)
	}
	if err := Validate(S{Rate: 3.0}); err == nil {
		t.Error("oneof float should fail")
	}
}

func TestOneOf_UnsupportedType(t *testing.T) {
	type S struct {
		Data bool `validate:"oneof=true false"`
	}
	// Bool is not supported by oneof, should pass (default case returns nil)
	if err := Validate(S{Data: true}); err != nil {
		t.Errorf("oneof unsupported should pass: %v", err)
	}
}

// ============================================================================
// Comparison Rules
// ============================================================================

func TestGt(t *testing.T) {
	type S struct {
		Val int `validate:"gt=10"`
	}
	if err := Validate(S{Val: 11}); err != nil {
		t.Errorf("gt pass: %v", err)
	}
	if err := Validate(S{Val: 10}); err == nil {
		t.Error("gt equal should fail")
	}
	if err := Validate(S{Val: 9}); err == nil {
		t.Error("gt less should fail")
	}
}

func TestGte(t *testing.T) {
	type S struct {
		Val int `validate:"gte=10"`
	}
	if err := Validate(S{Val: 10}); err != nil {
		t.Errorf("gte pass: %v", err)
	}
	if err := Validate(S{Val: 9}); err == nil {
		t.Error("gte should fail")
	}
}

func TestLt(t *testing.T) {
	type S struct {
		Val int `validate:"lt=10"`
	}
	if err := Validate(S{Val: 9}); err != nil {
		t.Errorf("lt pass: %v", err)
	}
	if err := Validate(S{Val: 10}); err == nil {
		t.Error("lt equal should fail")
	}
}

func TestLte(t *testing.T) {
	type S struct {
		Val int `validate:"lte=10"`
	}
	if err := Validate(S{Val: 10}); err != nil {
		t.Errorf("lte pass: %v", err)
	}
	if err := Validate(S{Val: 11}); err == nil {
		t.Error("lte should fail")
	}
}

func TestComparison_Uint(t *testing.T) {
	type S struct {
		Val uint `validate:"gt=5"`
	}
	if err := Validate(S{Val: 6}); err != nil {
		t.Errorf("gt uint pass: %v", err)
	}
	if err := Validate(S{Val: 5}); err == nil {
		t.Error("gt uint should fail")
	}
}

func TestComparison_Float(t *testing.T) {
	type S struct {
		Val float64 `validate:"lt=3.14"`
	}
	if err := Validate(S{Val: 3.13}); err != nil {
		t.Errorf("lt float pass: %v", err)
	}
	if err := Validate(S{Val: 3.14}); err == nil {
		t.Error("lt float should fail")
	}
}

func TestComparison_UnsupportedType(t *testing.T) {
	type S struct {
		Val string `validate:"gt=5"`
	}
	// String is not a numeric type, should pass (default returns nil)
	if err := Validate(S{Val: "abc"}); err != nil {
		t.Errorf("comparison on string should pass: %v", err)
	}
}

// ============================================================================
// String Rules
// ============================================================================

func TestContains(t *testing.T) {
	type S struct {
		URL string `validate:"contains=example"`
	}
	if err := Validate(S{URL: "https://example.com"}); err != nil {
		t.Errorf("contains pass: %v", err)
	}
	if err := Validate(S{URL: "https://other.com"}); err == nil {
		t.Error("contains should fail")
	}
	if err := Validate(S{URL: ""}); err != nil {
		t.Errorf("contains empty skip: %v", err)
	}
}

func TestContains_NonString(t *testing.T) {
	type S struct {
		Val int `validate:"contains=abc"`
	}
	if err := Validate(S{Val: 123}); err != nil {
		t.Errorf("contains non-string skip: %v", err)
	}
}

func TestStartsWith(t *testing.T) {
	type S struct {
		Path string `validate:"startswith=/api"`
	}
	if err := Validate(S{Path: "/api/users"}); err != nil {
		t.Errorf("startswith pass: %v", err)
	}
	if err := Validate(S{Path: "/v1/api"}); err == nil {
		t.Error("startswith should fail")
	}
	if err := Validate(S{Path: ""}); err != nil {
		t.Errorf("startswith empty skip: %v", err)
	}
}

func TestEndsWith(t *testing.T) {
	type S struct {
		File string `validate:"endswith=.go"`
	}
	if err := Validate(S{File: "main.go"}); err != nil {
		t.Errorf("endswith pass: %v", err)
	}
	if err := Validate(S{File: "main.py"}); err == nil {
		t.Error("endswith should fail")
	}
	if err := Validate(S{File: ""}); err != nil {
		t.Errorf("endswith empty skip: %v", err)
	}
}

func TestLowercase(t *testing.T) {
	type S struct {
		Key string `validate:"lowercase"`
	}
	if err := Validate(S{Key: "hello"}); err != nil {
		t.Errorf("lowercase pass: %v", err)
	}
	if err := Validate(S{Key: "Hello"}); err == nil {
		t.Error("lowercase should fail")
	}
	if err := Validate(S{Key: ""}); err != nil {
		t.Errorf("lowercase empty skip: %v", err)
	}
}

func TestUppercase(t *testing.T) {
	type S struct {
		Code string `validate:"uppercase"`
	}
	if err := Validate(S{Code: "ABC"}); err != nil {
		t.Errorf("uppercase pass: %v", err)
	}
	if err := Validate(S{Code: "Abc"}); err == nil {
		t.Error("uppercase should fail")
	}
	if err := Validate(S{Code: ""}); err != nil {
		t.Errorf("uppercase empty skip: %v", err)
	}
}

func TestExcludes(t *testing.T) {
	type S struct {
		Text string `validate:"excludes=script"`
	}
	if err := Validate(S{Text: "hello"}); err != nil {
		t.Errorf("excludes pass: %v", err)
	}
	if err := Validate(S{Text: "<script>alert</script>"}); err == nil {
		t.Error("excludes should fail")
	}
	if err := Validate(S{Text: ""}); err != nil {
		t.Errorf("excludes empty skip: %v", err)
	}
}

// ============================================================================
// Format Rules
// ============================================================================

func TestEmail(t *testing.T) {
	type S struct {
		Email string `validate:"email"`
	}
	if err := Validate(S{Email: "a@b.com"}); err != nil {
		t.Errorf("email pass: %v", err)
	}
	if err := Validate(S{Email: "invalid"}); err == nil {
		t.Error("email should fail")
	}
	if err := Validate(S{Email: ""}); err != nil {
		t.Errorf("email empty skip: %v", err)
	}
}

func TestEmail_NonString(t *testing.T) {
	type S struct {
		Val int `validate:"email"`
	}
	if err := Validate(S{Val: 123}); err != nil {
		t.Errorf("email non-string skip: %v", err)
	}
}

func TestURL(t *testing.T) {
	type S struct {
		URL string `validate:"url"`
	}
	if err := Validate(S{URL: "https://example.com"}); err != nil {
		t.Errorf("url pass: %v", err)
	}
	if err := Validate(S{URL: "not-a-url"}); err == nil {
		t.Error("url should fail")
	}
	if err := Validate(S{URL: ""}); err != nil {
		t.Errorf("url empty skip: %v", err)
	}
}

func TestNumeric(t *testing.T) {
	type S struct {
		Num string `validate:"numeric"`
	}
	if err := Validate(S{Num: "12345"}); err != nil {
		t.Errorf("numeric pass: %v", err)
	}
	if err := Validate(S{Num: "123abc"}); err == nil {
		t.Error("numeric should fail")
	}
	if err := Validate(S{Num: ""}); err != nil {
		t.Errorf("numeric empty skip: %v", err)
	}
}

func TestAlpha(t *testing.T) {
	type S struct {
		Name string `validate:"alpha"`
	}
	if err := Validate(S{Name: "hello"}); err != nil {
		t.Errorf("alpha pass: %v", err)
	}
	if err := Validate(S{Name: "hello123"}); err == nil {
		t.Error("alpha should fail")
	}
	if err := Validate(S{Name: ""}); err != nil {
		t.Errorf("alpha empty skip: %v", err)
	}
}

func TestAlphanumeric(t *testing.T) {
	type S struct {
		Code string `validate:"alphanumeric"`
	}
	if err := Validate(S{Code: "abc123"}); err != nil {
		t.Errorf("alphanumeric pass: %v", err)
	}
	if err := Validate(S{Code: "abc-123"}); err == nil {
		t.Error("alphanumeric should fail")
	}
	if err := Validate(S{Code: ""}); err != nil {
		t.Errorf("alphanumeric empty skip: %v", err)
	}
}

func TestUUID(t *testing.T) {
	type S struct {
		ID string `validate:"uuid"`
	}
	if err := Validate(S{ID: "550e8400-e29b-41d4-a716-446655440000"}); err != nil {
		t.Errorf("uuid pass: %v", err)
	}
	if err := Validate(S{ID: "not-a-uuid"}); err == nil {
		t.Error("uuid should fail")
	}
	if err := Validate(S{ID: ""}); err != nil {
		t.Errorf("uuid empty skip: %v", err)
	}
}

func TestHexColor(t *testing.T) {
	type S struct {
		Color string `validate:"hexcolor"`
	}
	tests := []struct {
		color   string
		wantErr bool
	}{
		{"#FFF", false},
		{"#FFFFFF", false},
		{"#aabbcc", false},
		{"#123", false},
		{"FFF", true},
		{"#GGG", true},
		{"#12345", true},
		{"", false},
	}
	for _, tt := range tests {
		err := Validate(S{Color: tt.color})
		if (err != nil) != tt.wantErr {
			t.Errorf("hexcolor(%q) err=%v, wantErr=%v", tt.color, err, tt.wantErr)
		}
	}
}

func TestDatetime_CustomFormat(t *testing.T) {
	type S struct {
		Date string `validate:"datetime=2006-01-02"`
	}
	if err := Validate(S{Date: "2024-01-15"}); err != nil {
		t.Errorf("datetime pass: %v", err)
	}
	if err := Validate(S{Date: "15/01/2024"}); err == nil {
		t.Error("datetime should fail")
	}
	if err := Validate(S{Date: ""}); err != nil {
		t.Errorf("datetime empty skip: %v", err)
	}
}

func TestDatetime_DefaultRFC3339(t *testing.T) {
	type S struct {
		Time string `validate:"datetime"`
	}
	if err := Validate(S{Time: "2024-01-15T10:30:00Z"}); err != nil {
		t.Errorf("datetime RFC3339 pass: %v", err)
	}
	if err := Validate(S{Time: "not-a-date"}); err == nil {
		t.Error("datetime RFC3339 should fail")
	}
}

func TestDatetime_NonString(t *testing.T) {
	type S struct {
		Val int `validate:"datetime"`
	}
	if err := Validate(S{Val: 123}); err != nil {
		t.Errorf("datetime non-string skip: %v", err)
	}
}

// ============================================================================
// Network Rules
// ============================================================================

func TestIP(t *testing.T) {
	type S struct {
		Addr string `validate:"ip"`
	}
	if err := Validate(S{Addr: "192.168.1.1"}); err != nil {
		t.Errorf("ip v4 pass: %v", err)
	}
	if err := Validate(S{Addr: "::1"}); err != nil {
		t.Errorf("ip v6 pass: %v", err)
	}
	if err := Validate(S{Addr: "not.an.ip"}); err == nil {
		t.Error("ip should fail")
	}
	if err := Validate(S{Addr: ""}); err != nil {
		t.Errorf("ip empty skip: %v", err)
	}
}

func TestIPv4(t *testing.T) {
	type S struct {
		Addr string `validate:"ipv4"`
	}
	if err := Validate(S{Addr: "10.0.0.1"}); err != nil {
		t.Errorf("ipv4 pass: %v", err)
	}
	if err := Validate(S{Addr: "::1"}); err == nil {
		t.Error("ipv4 should reject IPv6")
	}
	if err := Validate(S{Addr: "not-ip"}); err == nil {
		t.Error("ipv4 should reject invalid")
	}
	if err := Validate(S{Addr: ""}); err != nil {
		t.Errorf("ipv4 empty skip: %v", err)
	}
}

func TestIPv6(t *testing.T) {
	type S struct {
		Addr string `validate:"ipv6"`
	}
	if err := Validate(S{Addr: "2001:db8::1"}); err != nil {
		t.Errorf("ipv6 pass: %v", err)
	}
	if err := Validate(S{Addr: "192.168.1.1"}); err == nil {
		t.Error("ipv6 should reject IPv4")
	}
	if err := Validate(S{Addr: "not-ip"}); err == nil {
		t.Error("ipv6 should reject invalid")
	}
	if err := Validate(S{Addr: ""}); err != nil {
		t.Errorf("ipv6 empty skip: %v", err)
	}
}

// ============================================================================
// Collection Rules
// ============================================================================

func TestNotEmpty_Slice(t *testing.T) {
	type S struct {
		Items []int `validate:"notempty"`
	}
	if err := Validate(S{Items: []int{1}}); err != nil {
		t.Errorf("notempty pass: %v", err)
	}
	if err := Validate(S{Items: []int{}}); err == nil {
		t.Error("notempty empty should fail")
	}
	if err := Validate(S{Items: nil}); err == nil {
		t.Error("notempty nil should fail")
	}
}

func TestNotEmpty_Map(t *testing.T) {
	type S struct {
		M map[string]int `validate:"notempty"`
	}
	if err := Validate(S{M: map[string]int{"a": 1}}); err != nil {
		t.Errorf("notempty map pass: %v", err)
	}
	if err := Validate(S{M: map[string]int{}}); err == nil {
		t.Error("notempty empty map should fail")
	}
}

func TestNotEmpty_String(t *testing.T) {
	type S struct {
		Name string `validate:"notempty"`
	}
	if err := Validate(S{Name: "ok"}); err != nil {
		t.Errorf("notempty string pass: %v", err)
	}
	if err := Validate(S{Name: ""}); err == nil {
		t.Error("notempty empty string should fail")
	}
}

func TestUnique(t *testing.T) {
	type S struct {
		Tags []string `validate:"unique"`
	}
	if err := Validate(S{Tags: []string{"a", "b", "c"}}); err != nil {
		t.Errorf("unique pass: %v", err)
	}
	if err := Validate(S{Tags: []string{"a", "b", "a"}}); err == nil {
		t.Error("unique should fail for duplicates")
	}
}

func TestUnique_EdgeCases(t *testing.T) {
	type S struct {
		Tags []string `validate:"unique"`
	}
	// Nil should pass
	if err := Validate(S{Tags: nil}); err != nil {
		t.Errorf("unique nil pass: %v", err)
	}
	// Single element should pass
	if err := Validate(S{Tags: []string{"a"}}); err != nil {
		t.Errorf("unique single pass: %v", err)
	}
}

func TestUnique_Int(t *testing.T) {
	type S struct {
		IDs []int `validate:"unique"`
	}
	if err := Validate(S{IDs: []int{1, 2, 3}}); err != nil {
		t.Errorf("unique int pass: %v", err)
	}
	if err := Validate(S{IDs: []int{1, 2, 1}}); err == nil {
		t.Error("unique int should fail")
	}
}

func TestUnique_NonSlice(t *testing.T) {
	type S struct {
		Name string `validate:"unique"`
	}
	// Non-slice should pass
	if err := Validate(S{Name: "hello"}); err != nil {
		t.Errorf("unique non-slice pass: %v", err)
	}
}

// ============================================================================
// Factory edge cases
// ============================================================================

func TestIntFactory_EmptyParam(t *testing.T) {
	type S struct {
		Val int `validate:"min="`
	}
	// Empty param → noop
	if err := Validate(S{Val: 0}); err != nil {
		t.Errorf("min empty param should noop: %v", err)
	}
}

func TestIntFactory_InvalidParam(t *testing.T) {
	type S struct {
		Val int `validate:"min=abc"`
	}
	// Invalid int param → noop
	if err := Validate(S{Val: 0}); err != nil {
		t.Errorf("min invalid param should noop: %v", err)
	}
}

func TestFloatFactory_EmptyParam(t *testing.T) {
	type S struct {
		Val int `validate:"gt="`
	}
	if err := Validate(S{Val: 0}); err != nil {
		t.Errorf("gt empty param should noop: %v", err)
	}
}

func TestFloatFactory_InvalidParam(t *testing.T) {
	type S struct {
		Val int `validate:"gt=abc"`
	}
	if err := Validate(S{Val: 0}); err != nil {
		t.Errorf("gt invalid param should noop: %v", err)
	}
}

// ============================================================================
// Combined / Integration
// ============================================================================

func TestCombinedRules(t *testing.T) {
	type Config struct {
		Host     string   `validate:"required,ip"`
		Port     int      `validate:"required,min=1,max=65535"`
		Protocol string   `validate:"required,oneof=http https"`
		BasePath string   `validate:"required,startswith=/"`
		LogLevel string   `validate:"required,lowercase"`
		Tags     []string `validate:"notempty,unique,dive,required,min=1"`
	}

	valid := Config{
		Host:     "192.168.1.1",
		Port:     8080,
		Protocol: "https",
		BasePath: "/api",
		LogLevel: "debug",
		Tags:     []string{"production", "v2"},
	}
	if err := Validate(valid); err != nil {
		t.Fatalf("valid config should pass: %v", err)
	}

	invalid := Config{
		Host:     "not-an-ip",
		Port:     99999,
		Protocol: "ftp",
		BasePath: "api",
		LogLevel: "DEBUG",
		Tags:     []string{"a", "a"},
	}
	err := Validate(invalid)
	if err == nil {
		t.Fatal("invalid config should fail")
	}
	errs := err.(ValidationErrors)
	if len(errs) < 5 {
		t.Errorf("expected at least 5 errors, got %d: %v", len(errs), errs)
	}
}

// ============================================================================
// isZeroValue coverage
// ============================================================================

func TestIsZeroValue_AllTypes(t *testing.T) {
	type S struct {
		Str   string            `validate:"required"`
		Bool  bool              `validate:"required"`
		Int   int               `validate:"required"`
		Uint  uint              `validate:"required"`
		Float float64           `validate:"required"`
		Ptr   *int              `validate:"required"`
		Slice []int             `validate:"required"`
		Map   map[string]string `validate:"required"`
		Iface interface{}       `validate:"required"`
		Ch    chan int          `validate:"required"`
		Fn    func()            `validate:"required"`
	}

	// All zero → should all fail required
	err := Validate(S{})
	if err == nil {
		t.Fatal("all zero should fail")
	}
	errs := err.(ValidationErrors)
	// Should have errors for all fields
	if len(errs) < 10 {
		t.Errorf("expected at least 10 errors, got %d", len(errs))
	}

	// Non-zero struct (unusual type like chan)
	n := 42
	ch := make(chan int)
	fn := func() {}
	valid := S{
		Str: "a", Bool: true, Int: 1, Uint: 1, Float: 1.0,
		Ptr: &n, Slice: []int{1}, Map: map[string]string{"k": "v"},
		Iface: "x", Ch: ch, Fn: fn,
	}
	if err := Validate(valid); err != nil {
		t.Errorf("all non-zero should pass: %v", err)
	}
}
