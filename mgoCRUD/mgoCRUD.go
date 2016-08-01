package main

import (
	"fmt"
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Person struct {
	// See http://godoc.org/gopkg.in/mgo.v2/bson#Marshal
	// Define field name in the database if you want the name different from the struct.
	// If not defined, it will use lowercase name.
	Id    bson.ObjectId `bson:"_id,omitempty"`
	Name  string        `bson:"name"`
	Phone string
}

var (
	IsDrop = true
)

func main() {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	// Drop database.
	if IsDrop {
		err = session.DB("moon").DropDatabase()
		if err != nil {
			panic(err)
		}
	}

	// Collection `people`.
	c := session.DB("moon").C("people")
	result := Person{}

	// Index.
	index := mgo.Index{
		Key:        []string{"name", "phone"},
		Unique:     true,
		DropDups:   true,
		Background: true,
		Sparse:     true,
	}

	err = c.EnsureIndex(index)
	if err != nil {
		panic(err)
	}

	count1, err := c.Count()
	fmt.Println("<database drop | total peoples> :", count1)

	// Create.
	err = c.Insert(&Person{Name: "Ale", Phone: "+11 11 1111 1111"},
		&Person{Name: "Ale", Phone: "+22 22 2222 2222"},
		&Person{Name: "Cla", Phone: "+33 33 3333 3333"})
	if err != nil {
		log.Fatal(err)
	}

	count2, err := c.Count()
	fmt.Println("<database created | insert 3 peoples | total peoples> :", count2)

	// Query one.
	// See http://godoc.org/gopkg.in/mgo.v2#Collection.Find
	err = c.Find(bson.M{"name": "Ale"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("<query one | name: Ale> Phone :", result.Phone)

	// Query all [a].
	var results []Person
	err = c.Find(bson.M{"name": "Ale"}).Sort("-name").All(&results) // see http://godoc.org/gopkg.in/mgo.v2#Query.All
	if err != nil {
		panic(err)
	}

	fmt.Println()
	fmt.Println("<query all [a] | results>")
	fmt.Println(results)

	// Query all [b].
	iter := c.Find(nil).Limit(3).Iter()
	fmt.Println()
	fmt.Println("<query all [b] | results>")
	for iter.Next(&result) {
		fmt.Printf("results: %v\n", result.Name+" , "+result.Phone)
	}
	err = iter.Close()
	if err != nil {
		log.Fatal(err)
	}

	// Update.
	selector := bson.M{"name": "Cla"}
	updator := bson.M{"$set": bson.M{"phone": "+44 44 4444 4444"}}
	err = c.Update(selector, updator)
	if err != nil {
		panic(err)
	}

	err = c.Find(bson.M{"name": "Cla"}).One(&result)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println()
	fmt.Println("<updated | name : Ale> Phone :", result.Phone)

	// Remove.
	err = c.Remove(bson.M{"_id": result.Id})
	if err != nil {
		log.Fatal(err)
	}

	count3, err := c.Count()
	fmt.Println("<total peoples after delete> :", count3)
}
