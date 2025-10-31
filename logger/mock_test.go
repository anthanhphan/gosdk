package logger

// mockPrimitiveArrayEncoder is a mock implementation of zapcore.PrimitiveArrayEncoder
// to capture the encoded caller string for testing
type mockPrimitiveArrayEncoder struct {
	appendStringFunc func(string)
}

func (m *mockPrimitiveArrayEncoder) AppendString(s string) {
	if m.appendStringFunc != nil {
		m.appendStringFunc(s)
	}
}

func (m *mockPrimitiveArrayEncoder) AppendBool(bool)             {}
func (m *mockPrimitiveArrayEncoder) AppendByteString([]byte)     {}
func (m *mockPrimitiveArrayEncoder) AppendComplex128(complex128) {}
func (m *mockPrimitiveArrayEncoder) AppendComplex64(complex64)   {}
func (m *mockPrimitiveArrayEncoder) AppendFloat64(float64)       {}
func (m *mockPrimitiveArrayEncoder) AppendFloat32(float32)       {}
func (m *mockPrimitiveArrayEncoder) AppendInt(int)               {}
func (m *mockPrimitiveArrayEncoder) AppendInt8(int8)             {}
func (m *mockPrimitiveArrayEncoder) AppendInt16(int16)           {}
func (m *mockPrimitiveArrayEncoder) AppendInt32(int32)           {}
func (m *mockPrimitiveArrayEncoder) AppendInt64(int64)           {}
func (m *mockPrimitiveArrayEncoder) AppendUint(uint)             {}
func (m *mockPrimitiveArrayEncoder) AppendUint8(uint8)           {}
func (m *mockPrimitiveArrayEncoder) AppendUint16(uint16)         {}
func (m *mockPrimitiveArrayEncoder) AppendUint32(uint32)         {}
func (m *mockPrimitiveArrayEncoder) AppendUint64(uint64)         {}
func (m *mockPrimitiveArrayEncoder) AppendUintptr(uintptr)       {}
