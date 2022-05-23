package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/sbabiv/xml2map"
)

//------------------------------------------------------------------------------
// Here we compare slices of bytes
//------------------------------------------------------------------------------

// CompareBytes : comparing 2 slices of bytes containing the data for JSON or XML files
func CompareBytes(bytes1, bytes2 []byte, isXml bool, idProps map[string]string) (Comparison, error) {
	// if the XML option is activated, we compare 2 XML files
	if isXml {
		log.Print(xml2map.ErrInvalidDocument.Error())

		return nil, errors.New("XML is not handled yet!")
	}

	// handling the JSON unmarshalling
	var obj1 interface{}
	if errUnmarsh1 := json.Unmarshal(bytes1, &obj1); errUnmarsh1 != nil {
		return nil, fmt.Errorf("Error while unmarshalling the first data set. Cause: %s", errUnmarsh1)
	}

	var obj2 interface{}
	if errUnmarsh2 := json.Unmarshal(bytes2, &obj2); errUnmarsh2 != nil {
		return nil, fmt.Errorf("Error while unmarshalling the first data set. Cause: %s", errUnmarsh2)
	}

	// using the right comparison function, between 2 objects in general
	return compareObjects("", obj1, obj2, idProps)
}
