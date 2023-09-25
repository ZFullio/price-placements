package validation

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

const (
	MsgEmptyFeed string = "feed is empty"
)

type CustomInt64 struct {
	Int64 int64
	Valid bool
}

func (ci *CustomInt64) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	var s string
	if err := d.DecodeElement(&s, &start); err != nil {
		return err
	}

	if s == "undefined" {
		ci.Valid = false

		return nil
	}

	customI, err := strconv.Atoi(s)
	if err != nil {
		return err
	}

	ci.Int64 = int64(customI)

	return nil
}

func CheckString(path string, fieldName string, value string, results *[]string) (isOk bool) {
	if value == "" {
		*results = append(*results, fmt.Sprintf("field %s.%s is empty", path, fieldName))

		return false
	}

	return true
}

func CheckStringWithPos(idx int, path string, fieldName string, value string, results *[]string) (isOk bool) {
	if value == "" {
		*results = append(*results, fmt.Sprintf("field %s[%d].%s is empty", path, idx, fieldName))

		return false
	}

	return true
}

func CheckStringWithID(ID string, path string, fieldName string, value string, results *[]string) (isOk bool) {
	var idMessage string
	if ID == "" {
		idMessage = "InternalID not found"
	} else {
		idMessage = fmt.Sprintf("InternalID: %s", ID)
	}

	if value == "" {
		*results = append(*results, fmt.Sprintf("field %s.%s is empty. %s", path, fieldName, idMessage))

		return false
	}

	return true
}

func CheckZeroWithID[V int | float64 | float32](ID string, path string, fieldName string, value V, results *[]string) (isOk bool) {
	if value == 0 {
		*results = append(*results, fmt.Sprintf("field %s.%s is empty. InternalID: %s", path, fieldName, ID))

		return false
	}

	return true
}
