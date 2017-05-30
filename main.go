package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

var (
	words            []string
	config           Config
	dateFormatString string
)

const (
	TPL_REPLACE_STRING string = "%%"
)

// Init the RNG
func init() {
	rand.Seed(time.Now().UnixNano())
}

// Main entry point - do all the setup and get crackin'
func main() {
	startTime := time.Now()
	log.Println("Elastic Search Populator")

	confPath := "./config.json"
	if len(os.Args) > 1 {
		confPath = os.Args[1]
	}
	log.Printf("Reading config file at: %v\n", confPath)

	err := parseConfig(confPath)
	if err != nil {
		panic(err)
	}

	log.Println("Reading words dictionary")
	err = initWordsDict()
	if err != nil {
		panic(err)
	}

	log.Printf("Launching runner for %v entity types\n", len(config.Entities))

	// Begin the loop - each type of entity will get a goroutine and run
	// until complete.  This code will block on the wg.Wait() until all running
	// routines have marked the WG as done.
	wg := &sync.WaitGroup{}
	for _, entType := range config.Entities {
		wg.Add(1)
		go runEntityPopulation(entType, wg)
	}
	wg.Wait()

	endTime := time.Now()
	totalTime := endTime.Sub(startTime)
	log.Printf("All done - completed in %v seconds\n", totalTime.Seconds())
}

// For a given type of entity to be populated, generate the desired number of entries and then
// feed them via queue into a pool of workers.  Notify the wait group when all done
func runEntityPopulation(entConf ConfEntity, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Printf("Running population for entity type: %v\n", entConf.EsType)

	queue := make(chan Entry, config.QueueSize)
	stopChans := []chan bool{}
	idGen := NewIdGen()
	counter := NewCounter()

	// Are we running in safemode?
	doDelete := false
	if config.UnsafeIndexDelete {
		doDelete = true
	} else {
		if cliDeleteConfirm(entConf.Index) {
			doDelete = true
		}
	}

	if doDelete {
		log.Printf("Deleting previously existing index %v\n", entConf.Index)
		err := deleteIndex(entConf.Index)
		if err != nil {
			panic(err)
		}
	} else {
		log.Printf("Delete not confirmed for index: %v - skipping\n", entConf.Index)
		return
	}

	log.Printf("Starting %v workers for entity type %v\n", config.Workers, entConf.EsType)

	for i := 0; i < config.Workers; i++ {
		stopChan := make(chan bool)
		stopChans = append(stopChans, stopChan)
		go runWorker(queue, stopChan, counter)
	}

	log.Printf("Generating %v %v entries and placing them in queue\n", entConf.NumberEntities, entConf.EsType)
	for i := 0; i < entConf.NumberEntities; i++ {
		id := idGen.GetNext()
		queue <- generateEntry(entConf, id)
	}

	for {
		if counter.Count() >= entConf.NumberEntities {
			killWorkers(stopChans, entConf.Index)
			return
		}
	}

}

// Loop for a worker to run - it reads from the queue and writes the entries it gets
// there to elastic search, incrementing the shared counter for that entity type each time.
//
// It exits when it recieves the kill signal on the stop channel
func runWorker(queue chan Entry, stop chan bool, counter *Counter) {
	for {
		select {
		case entry := <-queue:
			err := writeEntry(entry)
			if err != nil {
				log.Printf("Entry type:%v id:%v got error: %v\n", entry.esType, entry.id, err.Error())
			} else if !config.QuietMode {
				log.Printf("Entry type:%v id:%v successfully posted\n", entry.esType, entry.id)
			}
			counter.Incr()
		case <-stop:
			log.Println("Working stopping")
			return
		}
	}
}

// Send the stop signal to all the workers so they can exit
func killWorkers(stopChans []chan bool, index string) {
	log.Printf("Killing workers for index: %v\n", index)
	for _, stopChan := range stopChans {
		stopChan <- true
	}
}

// ********************************
//      ES HTTP API FUNCS
// ********************************

// Delete an index from elastic search by name
func deleteIndex(indexName string) error {
	delUrl := fmt.Sprintf("%v/%v", config.BaseUrl, indexName)

	req, err := http.NewRequest("DELETE", fmt.Sprintf("%v/%v", config.BaseUrl, indexName), nil)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode == 404 {
		log.Printf("Index %v not found, skipping delete step\n", indexName)
	} else if resp.StatusCode != 200 && resp.StatusCode != 201 {
		log.Printf("DEL URL: %v\n", delUrl)
		return fmt.Errorf("ERROR: Got status %v on attempting to delete ES index", resp.StatusCode)
	}
	return nil
}

// Write an entry into the correct elastic search index
func writeEntry(e Entry) error {
	req, err := http.NewRequest("PUT", e.getESURI(config.BaseUrl), bytes.NewBuffer(e.toJSON()))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.StatusCode != 200 && resp.StatusCode != 201 {
		return fmt.Errorf("ERROR: Got status %v on attempting to write entry to ES", resp.StatusCode)
	}

	return nil
}

// ********************************
//      GLOBAL INIT FUNCS
// ********************************

// Parse configuration data
func parseConfig(path string) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return errors.New("Couldn't read config file: " + err.Error())
	}

	err = json.Unmarshal(raw, &config)
	if err != nil {
		return errors.New("Couldn't parse config JSON: " + err.Error())
	}

	dateFormatString = config.DateFormat

	return nil
}

// Read the words dictionary into an array
func initWordsDict() error {
	rawData, err := ioutil.ReadFile(config.DictFile)
	if err != nil {
		return nil
	}

	words = strings.Split(string(rawData), "\n")
	return nil
}

// ********************************
//          HELPER FUNCS
// ********************************

// Confirm via command prompt before deleting an index
func cliDeleteConfirm(idx string) bool {
	fmt.Printf("Index %v is about to be deleted, please confirm by typing 'yes' > ", idx)
	r := bufio.NewReader(os.Stdin)
	input, err := r.ReadString('\n')
	if err != nil {
		log.Printf("Error trying to read user input from CLI: %v\n", err.Error())
		return false
	}

	if strings.ToLower(strings.TrimSpace(input)) == "yes" {
		return true
	}
	return false
}
