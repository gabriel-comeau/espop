package main

// Config is a top level configuration object - it takes care of the global properties
// of the populator.
type Config struct {
	BaseUrl           string       `json:"esBaseUrl"`
	DictFile          string       `json:"dictFile"`
	QueueSize         int          `json:"queueSize"`
	Workers           int          `json:"workers"`
	TemplateBase      string       `json:"jsonTemplates"`
	DateFormat        string       `json:"dateFormat"`
	UnsafeIndexDelete bool         `json:"unsafeIndexDelete"`
	QuietMode         bool         `json:"quietMode"`
	Entities          []ConfEntity `json:"entities"`
}

// ConfEntity is an entity-specific configuration.  Each type of entity to be populated should have
// one of these entries.
type ConfEntity struct {
	Index          string      `json:"index"`
	EsType         string      `json:"esType"`
	DataFields     []DataField `json:"dataFields"`
	NumberEntities int         `json:"numberEntities"`
	Template       string      `json:"template"`
}

// DataField represents a given field of the entry's data.  Each piece of data must of a have
// a field name, type (string, int, float or date) and set the flag deciding if the data is to
// be randomized or not.  If not random, the associated "val" field should be set to the predefined
// data (for example, if the type is "string" and random is off, the "stringVal" field should be set
// with a predefined string to be used)
type DataField struct {
	Field       string `json:"field"`
	Type        string `json:"type"`
	RandomValue bool   `json:"randomValue"`

	// Random String settings
	RandomWordCount bool   `json:"randomWordCount"`
	MaxWordCount    int    `json:"maxWordCount"`
	Separator       string `json:"separator"`

	// Random Int settings
	MinIntVal int `json:"minIntVal"`
	MaxIntVal int `json:"maxIntVal"`

	// Random Float settings
	MinFloatVal float64 `json:"minFloatVal"`
	MaxFloatVal float64 `json:"maxFloatVal"`

	// Random Date settings
	MinDate string `json:"minDate"`
	MaxDate string `json:"maxDate"`

	// Fixed values
	StringVal string  `json:"stringVal"`
	IntVal    int     `json:"intVal"`
	FloatVal  float64 `json:"floatVal"`
	DateVal   string  `json:"dateVal"`
}
