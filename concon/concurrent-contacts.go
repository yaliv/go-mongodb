package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/yaliv/go-pkg/copydir"
)

func main() {
	// The data dir must have these files.
	contacts := map[string]bool{
		"contacts-au": true,
		"contacts-ca": true,
		"contacts-uk": true,
		"contacts-us": true,
	}

	fmt.Println(contacts)

	/* for key, val := range contacts {
		fmt.Print(key, ": ")
		fmt.Println(val)
	} */

	needDataSource := false

	// Read the data dir for existing files.
	existingFiles, err := ioutil.ReadDir("data")
	if err != nil {
		log.Println(err)
		needDataSource = true
	}
	for _, file := range existingFiles {
		name := strings.Split(file.Name(), ".csv")[0]
		if _, ok := contacts[name]; !ok {
			log.Printf("%q is not on the list.\n", file.Name())
			needDataSource = true
			break
		}
	}

	if needDataSource {
		if !supplyDataSource() {
			log.Fatalf("Failed to supply data source.")
		}
	}
}

func supplyDataSource() bool {
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		log.Println("GOPATH not found.")
		return false
	} else {
		dsLoc := gopath + "/src/github.com/yaliv/go-mongodb/concon/data"
		err := copydir.Copy(dsLoc, "data", true)
		if err != nil {
			log.Println(err)
			return false
		}
	}
	return true
}
