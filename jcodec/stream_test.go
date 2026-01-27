// Copyright (c) 2025 anthanhphan <an.thanhphan.work@gmail.com>

package jcodec

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
)

func TestEncoder(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)

	v := map[string]int{"a": 1}
	if err := enc.Encode(v); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	if !strings.Contains(buf.String(), `"a":1`) {
		t.Errorf("Expected JSON output containing \"a\":1, got %s", buf.String())
	}
}

func TestEncoder_SetIndent(t *testing.T) {
	var buf bytes.Buffer
	enc := NewEncoder(&buf)
	enc.SetIndent("", "  ")

	v := map[string]int{"a": 1}
	if err := enc.Encode(v); err != nil {
		t.Fatalf("Encode failed: %v", err)
	}

	expected := "{\n  \"a\": 1\n}\n"
	if buf.String() != expected {
		t.Errorf("Expected indented JSON:\n%q\nGot:\n%q", expected, buf.String())
	}
}

func TestDecoder(t *testing.T) {
	const jsonStream = `
		{"Name": "Ed", "Text": "Knock knock."}
		{"Name": "Sam", "Text": "Who's there?"}
		{"Name": "Ed", "Text": "Go fmt."}
		{"Name": "Sam", "Text": "Go fmt who?"}
		{"Name": "Ed", "Text": "Go fmt yourself!"}
	`
	type Message struct {
		Name, Text string
	}
	dec := NewDecoder(strings.NewReader(jsonStream))

	var messages []Message
	for {
		var m Message
		if err := dec.Decode(&m); err == io.EOF {
			break
		} else if err != nil {
			t.Fatalf("Decode failed: %v", err)
		}
		messages = append(messages, m)
	}

	if len(messages) != 5 {
		t.Errorf("Expected 5 messages, got %d", len(messages))
	}
	if messages[0].Name != "Ed" {
		t.Errorf("Expected first message from Ed, got %s", messages[0].Name)
	}
}

func TestRawMessage(t *testing.T) {
	type Color struct {
		Space string
		Point RawMessage // delay parsing until we know the color space
	}
	type RGB struct {
		R uint8
		G uint8
		B uint8
	}
	type YCbCr struct {
		Y  uint8
		Cb int8
		Cr int8
	}

	var j = []byte(`[
		{"Space": "YCbCr", "Point": {"Y": 255, "Cb": 0, "Cr": -10}},
		{"Space": "RGB",   "Point": {"R": 98, "G": 218, "B": 255}}
	]`)

	var colors []Color
	if err := Unmarshal(j, &colors); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if len(colors) != 2 {
		t.Fatalf("Expected 2 colors, got %d", len(colors))
	}

	for _, c := range colors {
		var dst interface{}
		switch c.Space {
		case "RGB":
			var rgb RGB
			if err := Unmarshal(c.Point, &rgb); err != nil {
				t.Fatalf("Unmarshal RGB failed: %v", err)
			}
			dst = rgb
			if rgb.R != 98 || rgb.G != 218 || rgb.B != 255 {
				t.Errorf("RGB mismatch: %+v", rgb)
			}
		case "YCbCr":
			var ycbcr YCbCr
			if err := Unmarshal(c.Point, &ycbcr); err != nil {
				t.Fatalf("Unmarshal YCbCr failed: %v", err)
			}
			dst = ycbcr
			if ycbcr.Y != 255 || ycbcr.Cb != 0 || ycbcr.Cr != -10 {
				t.Errorf("YCbCr mismatch: %+v", ycbcr)
			}
		}
		if dst == nil {
			t.Errorf("Unknown color space: %s", c.Space)
		}
	}

	// Verify RawMessage marshaling
	data, err := Marshal(colors[0])
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}
	if !strings.Contains(string(data), `"Space":"YCbCr"`) {
		t.Errorf("Marshal result missing expected field: %s", string(data))
	}
}

func TestRawMessage_MarshalJSON(t *testing.T) {
	var m RawMessage = []byte(`{"foo":"bar"}`)
	data, err := m.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON failed: %v", err)
	}
	if string(data) != `{"foo":"bar"}` {
		t.Errorf("Expected `{\"foo\":\"bar\"}`, got %s", string(data))
	}

	var nilM RawMessage
	data, err = nilM.MarshalJSON()
	if err != nil {
		t.Fatalf("MarshalJSON nil failed: %v", err)
	}
	if string(data) != "null" {
		t.Errorf("Expected null, got %s", string(data))
	}
}

func TestRawMessage_UnmarshalJSON(t *testing.T) {
	var m RawMessage
	if err := m.UnmarshalJSON([]byte(`{"a":1}`)); err != nil {
		t.Fatalf("UnmarshalJSON failed: %v", err)
	}
	if string(m) != `{"a":1}` {
		t.Errorf("Expected `{\"a\":1}`, got %s", string(m))
	}
}

func TestCompatibilityWithStdlib(_ *testing.T) {
	// Ensure our interfaces match stdlib
	var _ json.Marshaler = RawMessage{}
	var _ json.Unmarshaler = &RawMessage{}
}

func TestUtilityFunctions(t *testing.T) {
	// Test Compact
	var compactBuf bytes.Buffer
	src := []byte(`{
		"a": 1,
		"b": 2
	}`)
	if err := Compact(&compactBuf, src); err != nil {
		t.Fatalf("Compact failed: %v", err)
	}
	if compactBuf.String() != `{"a":1,"b":2}` {
		t.Errorf("Compact output mismatch: %s", compactBuf.String())
	}

	// Test Indent
	var indentBuf bytes.Buffer
	src = []byte(`{"a":1,"b":2}`)
	if err := Indent(&indentBuf, src, "", "  "); err != nil {
		t.Fatalf("Indent failed: %v", err)
	}
	expectedIndent := "{\n  \"a\": 1,\n  \"b\": 2\n}"
	if indentBuf.String() != expectedIndent {
		t.Errorf("Indent output mismatch:\nExpected:\n%s\nGot:\n%s", expectedIndent, indentBuf.String())
	}

	// Test HTMLEscape
	var htmlBuf bytes.Buffer
	src = []byte(`{"key":"<script>"}`)
	HTMLEscape(&htmlBuf, src)
	if !strings.Contains(htmlBuf.String(), `\u003cscript\u003e`) {
		t.Errorf("HTMLEscape output mismatch: %s", htmlBuf.String())
	}
}
