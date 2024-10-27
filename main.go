// main.go
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"eventstore_benchmark/slicestore"

	"github.com/fiatjaf/eventstore"
	"github.com/fiatjaf/eventstore/lmdb"
	"github.com/nbd-wtf/go-nostr"
)

func loadEvents(filename string) ([]*nostr.Event, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var events []*nostr.Event
	err = json.Unmarshal(data, &events)
	if err != nil {
		return nil, err
	}
	return events, nil
}

func benchmarkLoadEvents(events []*nostr.Event, store eventstore.Store) time.Duration {
	start := time.Now()
	for _, event := range events {
		store.SaveEvent(context.Background(), event)
	}
	return time.Since(start)
}

func benchmarkQueryEventsByFilter(store eventstore.Store, filter nostr.Filter) {
	ch, err := store.QueryEvents(context.Background(), filter)
	if err != nil {
		log.Fatalf("Failed to query events: %v", err)
	}
	for range ch {
		// Do nothing
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <eventstore_type>", os.Args[0])
	}

	load := false
	if len(os.Args) >= 2 {
		if os.Args[2] == "load" || os.Args[1] == "mem" {
			load = true
		} else {
			fmt.Println("Usage: ./main <eventstore_type>")
			fmt.Println("eventstore_type: mem | lmdb")
			os.Exit(0)
		}
	}

	eventstoreType := os.Args[1]

	events, err := loadEvents("events.json")
	if err != nil {
		log.Fatalf("Failed to load events: %v", err)
	}

	var store eventstore.Store
	switch eventstoreType {
	case "lmdb":
		store = &lmdb.LMDBBackend{
			Path:     "./data",
			MaxLimit: 1 << 30,
		}
	case "mem":
		store = &slicestore.SliceStore{}
	default:
		log.Fatalf("Unknown eventstore type: %s", eventstoreType)
	}

	store.Init()
	if load {
		fmt.Printf("Loading events...\n")
		loadDuration := benchmarkLoadEvents(events, store)
		fmt.Printf("Time to load events: %v\n", loadDuration)
	}

	filters := []nostr.Filter{
		nostr.Filter{
			Limit: 500,
		},
		nostr.Filter{
			Kinds: []int{1},
		},
		nostr.Filter{
			Kinds: []int{1, 5, 7},
		},
		// existing author
		nostr.Filter{
			Authors: []string{"76e6cc3224c036b4c090d8e76262d2e9db82cd748213a78c79fc62561f175a26"},
		},
		// non existing author
		nostr.Filter{
			Authors: []string{"00e6cc3224c036b4c090d8e76262d2e9db82cd748213a78c79fc62561f175a00"},
		},
		nostr.Filter{
			Authors: []string{
				"76e6cc3224c036b4c090d8e76262d2e9db82cd748213a78c79fc62561f175a26",
				"00e6cc3224c036b4c090d8e76262d2e9db82cd748213a78c79fc62561f175a00",
				"460c25e682fda7832b52d1f22d3d22b3176d972f60dcdc3212ed8c92ef85065c",
				"d6568480980ccdc2f7103e0d88120ef0e8a45b04aebb99b5564869f0553a78f6",
				"f0fb31d1810a9f95df3d178fcd67ca0b09879ad11e8689e56962cd839fb8ead4",
			},
		},
		// existing id
		nostr.Filter{
			IDs: []string{"a313b4fc15a63995fa1a3a99584cb32fb77c10e4e929bfbd94bb40053549f124"},
		},
		// non existing id
		nostr.Filter{
			IDs: []string{"0000000000063995fa1a3a99584cb32fb77c10e4e929bfbd94bb40053549f100"},
		},
		nostr.Filter{
			IDs: []string{
				"a313b4fc15a63995fa1a3a99584cb32fb77c10e4e929bfbd94bb40053549f124",
				"235e8dc15d2a20fa69515d11584fdf791bed1946399b82cc5ad9df8866faa6ac",
				"36cfbc6135deb4ca2536325b72da3d752fd8ab766fa09e4e71389fe677f8f5d6",
				"7022ad4f3f43fe424b50c11b671398abf1314776dfa0f474d506ea9f42be46c9",
				"cefdcd1452174d166b2533d23b708637e698711f1db1fddceaba5008e47b7803",
			},
		},
	}

	count := 1000
	for _, filter := range filters {
		start := time.Now()
		for i := 0; i < count; i++ {
			benchmarkQueryEventsByFilter(store, filter)
		}
		duration := time.Since(start)
		fmt.Printf("Filter: %v\n", filter)
		fmt.Printf("Average Time: %v\n", duration/time.Duration(count))
		fmt.Printf("Total Time: %v\n", duration)
	}
}
