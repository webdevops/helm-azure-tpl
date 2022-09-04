package azuretpl

import (
	"encoding/json"

	"github.com/microsoft/kiota-abstractions-go/serialization"
)

func (e *AzureTemplateExecutor) msGraphSerializeObject(resultObj serialization.Parsable) (obj interface{}, err error) {
	writer, err := e.msGraphClient.RequestAdapter().GetSerializationWriterFactory().GetSerializationWriter("application/json")
	if err != nil {
		return nil, err
	}

	err = writer.WriteObjectValue("", resultObj)
	if err != nil {
		return nil, err
	}

	serializedValue, err := writer.GetSerializedContent()
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(serializedValue, &obj)
	if err != nil {
		return nil, err
	}

	return
}
