package extension

import (
	"github.com/ChikyuKido/nande/exporter/extension"
	"reflect"
	"testing"
)

func TestSerialization(t *testing.T) {
	data := extension.Data{
		ExtensionName: "ExampleExtension",
		Interval:      15,
		Metrics: []extension.Metric{
			{Name: "Metric1", Data: int64(123), Type: extension.INT64},
			{Name: "Metric2", Data: true, Type: extension.BOOL},
			{Name: "Metric3", Data: 3.14, Type: extension.FLOAT64},
			{Name: "Metric4", Data: "Hello, World!", Type: extension.STRING},
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
