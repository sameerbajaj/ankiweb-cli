// Copyright 2026 Sameer Bajaj and contributors. Licensed under Apache-2.0. See LICENSE.

package ankiwebproto

import (
	"bytes"
	"testing"
)

func TestVarintEncoding(t *testing.T) {
	cases := []struct {
		val  uint64
		want []byte
	}{
		{0, []byte{0x00}},
		{1, []byte{0x01}},
		{127, []byte{0x7f}},
		{128, []byte{0x80, 0x01}},
		{300, []byte{0xac, 0x02}},
		{1731906299829, []byte{0xb5, 0xcf, 0xa3, 0xed, 0xb3, 0x32}},
	}

	for _, tc := range cases {
		got := EncodeVarint(tc.val)
		if !bytes.Equal(got, tc.want) {
			t.Errorf("EncodeVarint(%d) = %x, want %x", tc.val, got, tc.want)
		}
	}
}

func TestEncodeAndParsePB(t *testing.T) {
	// Construct a synthetic protobuf message:
	// Field 1: string "Hello"
	// Field 2: varint 42
	// Field 3: submessage with Field 1: string "World"
	var buf []byte
	buf = append(buf, EncodeString(1, "Hello")...)
	buf = append(buf, append(EncodeTag(2, 0), EncodeVarint(42)...)...)
	subMsg := EncodeString(1, "World")
	buf = append(buf, EncodeMessage(3, subMsg)...)

	fields := ParsePB(buf)
	if len(fields) != 3 {
		t.Fatalf("ParsePB returned %d fields, want 3", len(fields))
	}

	if fields[0].Num != 1 || string(fields[0].Bytes) != "Hello" {
		t.Errorf("field 0 = %+v, want string Hello", fields[0])
	}
	if fields[1].Num != 2 || fields[1].Varint != 42 {
		t.Errorf("field 1 = %+v, want varint 42", fields[1])
	}
	if fields[2].Num != 3 {
		t.Errorf("field 2 = %+v, want fieldNum 3", fields[2])
	}

	subFields := ParsePB(fields[2].Bytes)
	if len(subFields) != 1 || string(subFields[0].Bytes) != "World" {
		t.Errorf("subfield = %+v, want string World", subFields)
	}
}
