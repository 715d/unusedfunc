package main

import (
	"bytes"
	"io"
	"strings"

	"encoding/json/jsontext"
	json "encoding/json/v2"
)

type marshalDirect struct{}

func (marshalDirect) MarshalJSONTo(enc *jsontext.Encoder) error {
	marshalDirect{}.marshalHelper()
	return enc.WriteToken(jsontext.String("marshal-direct"))
}

func (marshalDirect) marshalHelper() {}

func (marshalDirect) Unused() {
	marshalDirect{}.unusedHelper()
}

func (marshalDirect) unusedHelper() {}

type marshalWriteNested struct{}

func (value *marshalWriteNested) MarshalJSON() ([]byte, error) {
	value.marshalHelper()
	return []byte(`"marshal-write"`), nil
}

func (*marshalWriteNested) marshalHelper() {}

type marshalEncodeNested struct{}

func (marshalEncodeNested) AppendText(b []byte) ([]byte, error) {
	marshalEncodeNested{}.marshalHelper()
	return append(b, "marshal-encode"...), nil
}

func (marshalEncodeNested) marshalHelper() {}

type marshalTextNested struct{}

func (marshalTextNested) MarshalText() ([]byte, error) {
	marshalTextNested{}.marshalHelper()
	return []byte("marshal-text"), nil
}

func (marshalTextNested) marshalHelper() {}

type unmarshalNested struct{}

func (value *unmarshalNested) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	value.unmarshalHelper()
	return dec.SkipValue()
}

func (*unmarshalNested) unmarshalHelper() {}

type unmarshalRead struct{}

func (value *unmarshalRead) UnmarshalJSON([]byte) error {
	value.unmarshalHelper()
	return nil
}

func (*unmarshalRead) unmarshalHelper() {}

type unmarshalDecodeNested struct{}

func (value *unmarshalDecodeNested) UnmarshalText([]byte) error {
	value.unmarshalHelper()
	return nil
}

func (*unmarshalDecodeNested) unmarshalHelper() {}

func main() {
	_, _ = json.Marshal(marshalDirect{})
	_ = json.MarshalWrite(io.Discard, struct {
		Value *marshalWriteNested
	}{Value: &marshalWriteNested{}})

	var encoded bytes.Buffer
	_ = json.MarshalEncode(jsontext.NewEncoder(&encoded), struct {
		AppendValue marshalEncodeNested
		TextValue   marshalTextNested
	}{})

	var unmarshaled struct {
		Value *unmarshalNested
	}
	_ = json.Unmarshal([]byte(`{"Value":{}}`), &unmarshaled)

	var read unmarshalRead
	_ = json.UnmarshalRead(strings.NewReader(`{}`), &read)

	var decoded struct {
		Value *unmarshalDecodeNested
	}
	_ = json.UnmarshalDecode(jsontext.NewDecoder(strings.NewReader(`{"Value":"text"}`)), &decoded)
}
