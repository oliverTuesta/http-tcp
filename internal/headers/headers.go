package headers

import (
	"bytes"
	"fmt"
)

type Headers map[string]string 


func NewHeaders() Headers {
	return Headers{}
}


var ERROR_BAD_FIELD_LINE_FORMAT = fmt.Errorf("bad field line format")
var ERROR_BAD_FIELD_LINE_NAME = fmt.Errorf("bad field line name")
var ERROR_BAD_FIELD_LINE_VALUE = fmt.Errorf("bad field line value")


func isTChar(b byte) bool {
	if b >= 'a' && b <= 'z' {
		return true
	}
	if b >= 'A' && b <= 'Z' {
		return true
	}

	if b >= '0' && b <= '9' {
		return true
	}

	switch b {
		case '!', '#', '$', '%', '&', '\'', '*',
		'+', '-', '.', '^', '_', '`', '|', '~':
		return true
	}

	return false
}

func formatFieldName(data []byte) []byte  {
	leftSpaces := 0
	for leftSpaces < len(data) && data[leftSpaces] == ' '{
		leftSpaces++
	}
	data = data[leftSpaces:]
	for _, c := range data {
		if !isTChar(c) {
			return nil
		}
	}

	if len(data) == 0 {
		return nil
	}

	return bytes.ToLower(data)
}

func formatFieldValue(data []byte) []byte {
	i := 0
	for i < len(data) && data[i] == ' '{
		i++
	}
	data = data[i:]

	i = len(data)
	for i >= 0 && data[i-1] == ' '{
		i--
	}
	data = data[:i]

	if len(data) == 0 {
		return nil
	}

	return data
}

var HEADER_SEPARATOR = []byte("\r\n")
var LINE_SEPARATOR = []byte(":")

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	idx := bytes.Index(data, HEADER_SEPARATOR)
	if idx == -1 {
		return 0, false, nil
	}

	n = idx + len(HEADER_SEPARATOR)
	if idx == 0 {
		return n, true, nil
	}

	line := data[:idx]

	idxSeparator := bytes.Index(line, LINE_SEPARATOR)
	if idxSeparator == -1 {
		return 0, false, ERROR_BAD_FIELD_LINE_FORMAT
	}

	fieldName := formatFieldName(line[:idxSeparator])
	if fieldName == nil {
		return 0, false, ERROR_BAD_FIELD_LINE_NAME
	}

	fieldValue := formatFieldValue(line[idxSeparator+len(LINE_SEPARATOR):])
	if fieldValue == nil {
		return 0, false, ERROR_BAD_FIELD_LINE_VALUE
	}

	strName := string(fieldName)
	strValue := string(fieldValue)

	curr, exists := h[strName]

	if !exists {
		h[strName] = strValue	
	} else {
		h[strName] = curr + ", " + strValue	
	}

	return n, false, nil
}
