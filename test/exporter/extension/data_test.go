package extension

import (
	"github.com/ChikyuKido/nande/exporter/extension"
	"reflect"
	"testing"
)

func TestSerialization(t *testing.T) {
	data := extension.Data{
		ExtensionName:  "ExampleExtension",
		Interval:       15,
		ProcessingTime: 10,
		Metrics: []string{
			"test", "test2",
		},
	}

	serializedData, err := extension.SerializeData(data)
	if err != nil {
		t.Fatalf("serialize data failed: %v", err)
		return
	}
	deserializedData, err := extension.DeserializeData(serializedData)
	if err != nil {
		t.Fatalf("deserialize data failed: %v", err)
		return
	}

	if !reflect.DeepEqual(data, deserializedData) {
		t.Fatalf("data and deserialized are not equal")
	}
}
