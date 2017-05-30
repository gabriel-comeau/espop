package main

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// Entry represents a piece of data to be written to a specific index in elastic search
type Entry struct {
	id      int
	esIndex string
	esType  string
	tpl     string
	data    []EntryData
}

// EntryData represents the stored data to be printed
type EntryData struct {
	key   string
	value interface{}
}

// Figure out the URI used in posting this entry to ES
func (e Entry) getESURI(base string) string {
	return fmt.Sprintf("%v/%v/%v/%v?pretty", base, e.esIndex, e.esType, e.id)
}

// Convert a piece of an entry's data to JSON suitable string
func (ed EntryData) getStringValue() string {

	switch val := ed.value.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%v", val)
	case float64:
		return fmt.Sprintf("%v", val)
	case time.Time:
		return fmt.Sprintf("%v", val.Format(dateFormatString))
	default:
		log.Println("Wrong data contained in EntryData can't convert properly -- skipping")
		return ""
	}
}

// Convert to a JSON entry capable of being written to ES
func (e Entry) toJSON() []byte {
	json := e.tpl
	for _, df := range e.data {
		replKey := fmt.Sprintf("%v%v%v", TPL_REPLACE_STRING, df.key, TPL_REPLACE_STRING)
		json = strings.Replace(json, replKey, df.getStringValue(), -1)
	}
	return []byte(json)
}
