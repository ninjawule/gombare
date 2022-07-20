package core

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/sbabiv/xml2map"
)

//------------------------------------------------------------------------------
// Here we compare slices of bytes
//------------------------------------------------------------------------------

// compareBytes : comparing 2 slices of bytes containing the data for JSON or XML files
func compareBytes(bytes1, bytes2 []byte, options *ComparisonOptions, doLog bool) (Comparison, error) {
	// if the XML option is activated, we compare 2 XML files
	if options.GetFileType() == FileTypeXML {
		// handling the XML unmarshalling
		if doLog {
			options.logger.Info("Unmarshalling the first file")
		}

		map1, err1 := xml2map.NewDecoder(bytes.NewReader(bytes1)).Decode()
		if err1 != nil {
			return nil, fmt.Errorf("Error while unmarshalling the first XML data set. Cause: %s", err1)
		}

		if doLog {
			options.logger.Info("Unmarshalling the second file")
		}

		map2, err2 := xml2map.NewDecoder(bytes.NewReader(bytes2)).Decode()
		if err2 != nil {
			return nil, fmt.Errorf("Error while unmarshalling the second XML data set. Cause: %s", err2)
		}

		if doLog {
			options.logger.Info("Done unmarshalling the two files")
		}

		// using the right comparison function, between 2 objects in general
		return compareMaps(options.idParams, map1, map2, options, "", false)
	}

	// handling the JSON unmarshalling
	var obj1 interface{}

	if doLog {
		options.logger.Info("Unmarshalling the first file")
	}

	if errUnmarsh1 := json.Unmarshal(bytes1, &obj1); errUnmarsh1 != nil {
		return nil, fmt.Errorf("Error while unmarshalling the first JSON data set. Cause: %s", errUnmarsh1)
	}

	var obj2 interface{}

	if doLog {
		options.logger.Info("Unmarshalling the second file")
	}

	if errUnmarsh2 := json.Unmarshal(bytes2, &obj2); errUnmarsh2 != nil {
		return nil, fmt.Errorf("Error while unmarshalling the second JSON data set. Cause: %s", errUnmarsh2)
	}

	if doLog {
		options.logger.Info("Done unmarshalling the two files")
	}

	// using the right comparison function, between 2 objects in general
	return compareObjects(nil, nil, options.idParams, obj1, obj2, options, "")
}
