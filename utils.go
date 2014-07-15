package main

import (
	"encoding/json"
	"encoding/xml"
	"strings"
)

func parseJsonData(msg string) (map[string]interface{}, *errorHandler) {
	var i map[string]interface{}
	c := []byte(msg)

	err := json.Unmarshal(c, &i)
	if err != nil {
		return nil, &errorHandler{err, "Unable to un-marshal data", 100056}
	}

	return i, nil
}

/* Get xml element value */
func getXmlData(xml_data []byte, key string) string {
	ret_str := ""
	rr := strings.NewReader(string(xml_data))
	decoder := xml.NewDecoder(rr)
	for {
		token, _ := decoder.Token()
		if token == nil {
			break
		}
		switch t := token.(type) {
		case xml.StartElement:
			elmt := xml.StartElement(t)
			for _, v := range elmt.Attr {
				if v.Name.Local == key {
					ret_str = v.Value
				}
			}
		}
	}
	return ret_str
}
