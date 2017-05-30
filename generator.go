package main

import (
	"io/ioutil"
	"log"
	"math/rand"
	"strings"
	"time"
)

// Generate a random entry based on the configuration-defined entry
func generateEntry(entConf ConfEntity, id int) Entry {

	dataArr := make([]EntryData, 0)
	for _, df := range entConf.DataFields {
		dataArr = append(dataArr, createDataEntry(df, entConf.EsType))
	}

	templateData := readTemplateData(entConf.Template)

	return Entry{
		id:      id,
		esIndex: entConf.Index,
		esType:  entConf.EsType,
		tpl:     templateData,
		data:    dataArr,
	}
}

// Read the data from the template file provided by the entity config
func readTemplateData(path string) string {
	raw, err := ioutil.ReadFile(config.TemplateBase + "/" + path)
	if err != nil {
		panic(err)
	}

	return string(raw)
}

// Create the entry data
func createDataEntry(df DataField, esType string) EntryData {

	ed := EntryData{
		key: df.Field,
	}

	switch strings.ToLower(df.Type) {

	case "string":
		ed.value = getStringVal(df, esType)
	case "int":
		ed.value = getIntVal(df, esType)
	case "float":
		ed.value = getFloatVal(df, esType)
	case "date":
		ed.value = getDateVal(df, esType)
	default:
		log.Printf("Can't use data field of type: %v, skipping", df.Type)
	}

	return ed
}

// Generate a string val - either random or defined
func getStringVal(df DataField, esType string) string {
	if df.RandomValue {
		return getRandomText(df.RandomWordCount, df.MaxWordCount, df.Separator)
	}
	val := df.StringVal
	if val == "" {
		log.Printf("WARNING: datafield %v for entity %v was set to non random string but had no default value\n", df.Field, esType)
	}
	return val
}

// Generate an int val - random or defined.
func getIntVal(df DataField, esType string) int {
	if df.RandomValue {
		return getRandomInt(df.MaxIntVal)
	}
	val := df.IntVal
	if val == 0 {
		log.Printf("WARNING: Datafield %v for entity %v was set to non-random int but had no default value\n", df.Field, esType)
	}
	return val
}

// Generate a float val - either random or defined
func getFloatVal(df DataField, esType string) float64 {
	if df.RandomValue {
		return getRandomFloat(df.MaxFloatVal)
	}
	val := df.FloatVal
	if val == 0 {
		log.Printf("WARNING: Datafield %v for entity %v was set to non-random float but had no default value\n", df.Field, esType)
	}
	return val
}

// Generate a date value - either random or defined.  If random, it will require a min/max defined in the entity conf.
func getDateVal(df DataField, esType string) time.Time {
	if df.RandomValue {

		// First need to get min/max dates parsed and into unixtimes so we can randomize the number of seconds between them
		minDate, err := time.Parse(dateFormatString, df.MinDate)
		if err != nil {
			log.Printf("ERROR: minDate value was not parseable for field %v in entity %v - defaulting to current time\n", df.Field, esType)
			return time.Now()
		}

		maxDate, err := time.Parse(dateFormatString, df.MaxDate)
		if err != nil {
			log.Printf("ERROR: maxDate value was not parseable for field %v in entity %v - defaulting to current time\n", df.Field, esType)
			return time.Now()
		}

		diffSeconds := maxDate.Unix() - minDate.Unix()
		randSeconds := minDate.Unix() + int64(getRandomInt(int(diffSeconds)))

		randDate := time.Unix(randSeconds, 0)
		return randDate

	}
	strVal := df.DateVal
	date, err := time.Parse(dateFormatString, strVal)
	if err != nil {
		log.Printf("ERROR: Date type Datafield %v for entity %v was not parseable - defaulting to current time\n", df.Field, esType)
		return time.Now()
	}

	return date
}

// Generate a random string of text WC words long, each word separated by separator
//
// If the randWc flag is set, the number of words will be between 1 and wc (wc becoming the "max")
func getRandomText(randWc bool, wc int, separator string) string {
	ret := ""

	if randWc {
		wc = rand.Intn(wc-1) + 1
	}

	for i := 0; i < wc; i++ {
		ret += strings.Title(words[rand.Intn(len(words)-1)])
		if i < wc-1 {
			ret += separator
		}
	}
	return ret
}

// Generate a random floating point number - note this isn't perfect and max not behave
// quite as you desire because of integer truncating
func getRandomFloat(max float64) float64 {
	return float64(rand.Intn(int(max))) + rand.Float64()
}

// Generate a random integer between 0 and max
func getRandomInt(max int) int {
	return rand.Intn(max)
}
