# ESPOP - Elasticsearch Testdata Populator

This program is designed to generate test data for [elasticsearch](https://www.elastic.co/products/elasticsearch).  The data itself is based on templates so you can create any kind of document that you want.

## Installation

Espop is written in [Go](https://golang.org).  You'll need a working Go install on your system to compile and install it.  Following the [installation instructions](https://golang.org/doc/install) and then the [workspace setup directions](https://golang.org/doc/code.html#Workspaces) will get you ready to go.  After you can follow the directions below:

1. Clone the project: `git clone https://github.com/gabriel-comeau/espop.git`
2. Switch to the directory of the cloned project
3. Compile-only (binary in project dir) with `go build` OR:
4. Install to $GOPATH/bin folder: `go install`

Once this is done, you can read the **configuration** section to get started.

## Dependencies

### Build time

Nothing outside of the go stdlib.

### Run time

A working elasticsearch instance (tested against **5.2** and **5.3** versions)

## Configuration

There are two levels to configuring ESPOP: configuring the behavior of the program itself, and configuring the entities you will be writing into elasticsearch.  Additionally,
those entities will each require a template file.

### Top-level program config

Here's an example configuration file:

```javascript
{
    "esBaseUrl": "http://localhost:9200",
    "dictFile": "/usr/share/dict/words",
    "jsonTemplates": "templates",
    "queueSize" :10000,
    "workers": 16,
    "dateFormat": "2006-01-02T15:04:05-07:00",
    "entities": [ ... ]
}
```

* **esBaseUrl** - This is the base URI for the elasticsearch instance you will be populating
* **dictFile** - Path to dictionary file, which is a newline separated entry of words, used in generating random strings
* **jsonTemplates** - Path to folder where entity templates will be stored
* **queueSize** - Each entity type has a work queue read by multiple workers - this defines how big it can get before blocking the producer
* **workers** - Each entity type has 1 or more worker goroutines pulling items from the queue and writing them to elastic.  This defines how many per entity type.
* **dateFormat** - The format that dates will be read (from entity config) and written (into templates).  The specific value of the date is [defined here](https://golang.org/pkg/time/#pkg-constants)
* **entities** - The array of individual entities which will be populated

### Entity config

Each entity is defined as a JSON object and is expected to match up to a template.  You need to define to entity-wide data (such as the ES index name) and then define the individual fields that will be substituted out of the template with their generated value.  Values have types and can be randomized (with randomization rules depending on the data type) or have their values specified directly.

Here's an example entity config for an invented "checkin" type event.  The template for this config will be in the next section so looking at both will probably make the most sense:

```javascript
{
    "index": "checkins",
    "esType": "checkin",
    "template": "checkin.json.tpl",
    "numberEntities": 50000,
    "dataFields": [
        {
            "field": "location_name",
            "type": "string",
            "randomValue": true,
            "randomWordCount": true,
            "maxWordCount": 3,
            "separator": " "
        },
        {
            "field": "checkin_user_id",
            "type": "int",
            "randomValue": true,
            "minIntVal": 1,
            "maxIntVal": 10000
        },
        {
            "field": "longitude",
            "type": "float",
            "randomValue": true,
            "minFloatVal": -180,
            "maxFloatVal": 180
        },
        {
            "field": "latitude",
            "type": "float",
            "randomValue": true,
            "minFloatVal": -90,
            "maxFloatVal": 90
        },
        {
            "field": "checkin_date",
            "type": "date",
            "randomValue": true,
            "minDate": "2017-04-01T00:00:00-04:00",
            "maxDate": "2017-04-30T23:59:59-04:00"
        },
        {
            "field": "checkin_type",
            "type": "int",
            "randomValue": false,
            "intVal": 1
        }
    ]
}
```

#### Entity-wide config

* **index** - The name of the elasticsearch index these entities will be written into
* **esType** - The name of the document type elastic search will use for these entities
* **template** - The name of the template file for these entities (template should be located in the globally defined templates folder)
* **numberEntites** - How many of this kind of entity should be generated and written to the index
* **dataFields** - The definition of each individual data type

#### Datafield definitions (all typeS)

* **field** - The name of the field.  Note, this is the name of field for substitution - if this is set to "foo" and your template has a %%foo%% in it, that's where this generated value will be put.
* **type** - The data type of the field.  Current possible data types are *int*, *float*, *string* and *date*.
* **randomValue** - Boolean flag indicating whether this value should be randomly generated or use a directly specified value

##### String-type definitions

Note that random strings are generated using the provided dictionary file (defined in top level app config), so total character length can't be guaranteed unless using a dictionary file containing only words of the same length.

* **stringVal** - If the randomValue flag is set to false, this field should be present and have its value set to what you want appearing in the generated docs for this field.
* **randomWordCount** - If this is set, randomly generated strings will not all contain the same number of words, the word count will also be randomly generated.
* **maxWordCount** - Upper limit of word count if using the randomWordCount flag.  If not, this will be the explicit word count.
* **separator** - String to be inserted between words.  Usually a single space character is what is desired here, but you might also want dashes (to generated slugs) or an empty string.

##### Int-type definitions

* **intVal** - Explicit value to use for an int if the randomValue flag is set to false.
* **minIntVal** - Smallest randomly generated int possible
* **maxIntVal** - Largest randomly generated int possible

##### Float-type definitions

* **floatVal** - Explicit value to use for a float if the randomValue flag is set to false.
* **minFloatVal** - Smallest randomly generated float possible
* **maxFloatVal** - Largest randomly generated float possible

##### Date-type definitions

Note that the format of the dates used in the min/max date fields must match the format of the date specified in the global part of the config.  Randomized dates are randomized to 1 second resolution.

* **dateVal** - Explicit value to use for a date if the randomValue flag is set to false.
* **minDate** - Smallest randomly generated date possible.
* **maxFloatVal** - Largest randomly generated date possible


### Templates

Templates are very simple - they are just plain text files with marked values which get substitued with the datafield values as defined above.  Here's an example template for the *checkin* event defined in the previous section:

```javascript
{
    "location_name": "%%location_name%%",
    "user_id": %%checkin_user_id%%,
    "geolocation": {
      "long": %%longitude%%,
      "lat": %%latitude%%
    },
    "checkin_date": "%%checkin_date%%",
    "checkin_type": %%checkin_type%%
}
```

As you can see, the structure doesn't really matter - technically this doesn't even need to be a JSON file but Elasticsearch expects documents in JSON format so it won't be very happy with you if you were to do that.  The field-names and the substitution-value-name don't need to match, but they can of course.  Anything that will be a string or date should be wrapped in quotes and the numeric int/float values shouldn't be.

## Possible improvements

Here are a few things I think that could be done to improve the program - in my own use I haven't needed them yet but they're things to think about.

* Variable length array support - there's no way to make lists of stuff that aren't fixed length.  Not sure how to do this without massively increasing complexity.
* Some of the names of items in the datafield config are needlessly specific: using interface{} in the code, the intVal, stringVal, floatVal and dateVals could just be "val".

## License

MIT.  See COPYING file.