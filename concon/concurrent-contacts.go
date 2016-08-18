package main

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/ctessum/macreader"
	"github.com/yaliv/go-pkg/copydir"
	"gopkg.in/mgo.v2"
)

const (
	// Data source dir.
	dsDir = "data"
	// Database name.
	dbName = "banana"
)

// The data source dir must have these files.
var contactFiles = map[string]bool{
	"contacts-au": true,
	"contacts-ca": true,
	"contacts-uk": true,
	"contacts-us": true,
}

func main() {
	// Read the data source dir for existing files.
	// Then supply it if needed.
	if needDataSource() {
		if !supplyDataSource() {
			log.Fatal("Failed to supply data source.")
		}
	}

	// Let's move to MongoDB processing.
	mongoJob()
}

func isFatal(message string, err error) {
	if err != nil {
		log.Fatal(message, err)
	}
}

func needDataSource() bool {
	existingFiles, err := ioutil.ReadDir(dsDir)
	if err != nil {
		log.Println(err)
		return true
	}
	for _, file := range existingFiles {
		name := strings.Split(file.Name(), ".csv")[0]
		if _, ok := contactFiles[name]; !ok {
			log.Printf("%q is not on the list.\n", file.Name())
			return true
		}
	}
	return false
}

func supplyDataSource() bool {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Println("GOPATH not found.")
		return false
	} else {
		dsMaster := gopath + "/src/github.com/yaliv/go-mongodb/concon/data"
		err := copydir.Copy(dsMaster, dsDir, true)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}

func mongoJob() {
	// Create a session which maintains a pool of socket connections
	// to our MongoDB.
	moss, err := mgo.Dial("localhost")
	isFatal("Create Session: ", err)

	// Set session mode.
	moss.SetMode(mgo.Monotonic, true)

	// Create a wait group to manage the Goroutines.
	var wg sync.WaitGroup
	wg.Add(len(contactFiles))

	for contactFile := range contactFiles {
		go contactCreate(contactFile, &wg, moss)
	}

	// Wait for all the queries to complete.
	wg.Wait()
	log.Println("All Queries Completed.")
}

func contactCreate(contactFile string, wg *sync.WaitGroup, moss *mgo.Session) {
	// Decrement the wait group count so the program knows this
	// has been completed once the Goroutine exits.
	defer wg.Done()

	// Request a socket connection from the session to process our query.
	// Close the session when the Goroutine exits and put the connection back
	// into the pool.
	sessCopy := moss.Copy()
	defer sessCopy.Close()

	// Get a collection to execute the query against.
	collection := sessCopy.DB(dbName).C(contactFile)

	// If the records are not empty, do not continue.
	n, err := collection.Count()
	isFatal("Count existing records: ", err)
	if n > 0 {
		return
	}

	// Open CSV.
	file, err := os.Open(dsDir + "/" + contactFile + ".csv")
	isFatal("Open CSV: ", err)
	defer file.Close()

	// Read CSV, filtered with a CR to LF converter.
	r := csv.NewReader(macreader.New(file))

	// We need to know the field names.
	fields, err := r.Read()
	isFatal("Read field names: ", err)

	// Prepare a map to accommodate the flexible data structure.
	contactMap := make(map[string]string)

	// First record.
	first, err := r.Read()
	isFatal("Read 1st: ", err)
	fmt.Printf("%s: %s\n", contactFile, first)

	// Cache record values to the map.
	for i := 0; i < len(fields); i++ {
		contactMap[fields[i]] = first[i]
	}
	fmt.Printf("%s: %s\n", contactFile, contactMap)

	// Perform Insert.
	err = collection.Insert(contactMap)
	isFatal("Insert 1st: ", err)
}
