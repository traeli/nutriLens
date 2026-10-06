package pgsql

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestJSONListRoundTrip(t *testing.T) {
	want := JSONList[uint]{3, 7, 11}
	value, err := want.Value()
	if err != nil {
		t.Fatalf("Value() error = %v", err)
	}
	var got JSONList[uint]
	if err := got.Scan(value); err != nil {
		t.Fatalf("Scan() error = %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("round trip = %#v, want %#v", got, want)
	}
}

func TestJSONListMarshalsNilAsArray(t *testing.T) {
	data, err := json.Marshal(JSONList[uint](nil))
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if string(data) != "[]" {
		t.Fatalf("Marshal(nil) = %s, want []", data)
	}
}
