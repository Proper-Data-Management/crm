package utils

import "strings"
import xj "github.com/basgys/goxml2json"

//Comment XmltoJSONString
func XmltoJSONString(data string) (string, error) {
	//Преобразование XML строки в JSON строку

	xml := strings.NewReader(data)
	json, err := xj.Convert(xml)
	if err != nil {
		return "", err
	}

	return json.String(), nil

}
