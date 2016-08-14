package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/yaliv/go-pkg/copydir"
)

var (
	// Data source dir.
	dsDir = "data"
	// The data source dir must have these files.
	contacts = map[string]bool{
		"contacts-au": true,
		"contacts-ca": true,
		"contacts-uk": true,
		"contacts-us": true,
	}
)

func main() {
	fmt.Println(contacts)

	/* for key, val := range contacts {
		fmt.Print(key, ": ")
		fmt.Println(val)
	} */

	// Read the data dir for existing files.
	// Then supply it if needed.
	if needDataSource() {
		if !supplyDataSource() {
			log.Fatalf("Failed to supply data source.")
		}
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
		if _, ok := contacts[name]; !ok {
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
