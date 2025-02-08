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

	badger_v4 "github.com/dgraph-io/badger/v4"
	"github.com/fiatjaf/eventstore"
	"github.com/fiatjaf/eventstore/badger"

	// "github.com/fiatjaf/eventstore/lmdb"
	"github.com/fiatjaf/eventstore/postgresql"
	"github.com/fiatjaf/eventstore/sqlite3"
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
	for i, event := range events {
		if i%100 == 0 {
			fmt.Printf(".")
			if i%8000 == 0 && i != 0 {
				fmt.Printf("\n")
			}
		}
		store.SaveEvent(context.Background(), event)
	}
	return time.Since(start)
}

func benchmarkQueryEventsByFilter(store eventstore.Store, filter nostr.Filter) {
	ch, err := store.QueryEvents(context.Background(), filter)
	if err != nil {
		log.Fatalf("Failed to query events: %v", err)
	}
	for c := range ch {
		//fmt.Println(c)
		// just force checking if it is really loaded, so the compiler do not eliminate the code
		if c.ID == "" {
			log.Fatalf("Invalid event: %v", c)
		}
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %s <eventstore_type>", os.Args[0])
	}

	load := false
	if len(os.Args) > 2 {
		if os.Args[2] == "load" || os.Args[1] == "mem" || os.Args[1] == "mem_sqlite" {
			load = true
		} else {
			fmt.Println("Usage: ./main <eventstore_type> [load]")
			fmt.Println("eventstore_type: mem | lmdb | badger | sqlite | mem_sqlite")
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
	// case "lmdb":
	// 	store = &lmdb.LMDBBackend{
	// 		Path:     "./data",
	// 		MapSize:  1 << 30,
	// 		MaxLimit: 500,
	// 	}
	case "badger":
		store = &badger.BadgerBackend{
			Path:     "./data",
			MaxLimit: 500,
		}
	case "badgerlowmem":
		store = &badger.BadgerBackend{
			Path:     "./data",
			MaxLimit: 500,
			BadgerOptionsModifier: func(opts badger_v4.Options) badger_v4.Options {
				opts.
					// 1) Make memtable smaller => less RAM usage, but more frequent flushes
					WithMemTableSize(32 << 20). // default is 64 MB
					// 2) Reduce the number of concurrent memtables => lowers peak memory usage
					WithNumMemtables(2). // default is 5
					// 3) Lower block cache => less memory used for caching compressed blocks
					WithBlockCacheSize(32 << 20). // default is 256 MB
					// 4) Set a modest index cache => reduce memory used by bloom filters / indexes
					WithIndexCacheSize(8 << 20). // default is 0 => all indices in memory
					// 5) Pick a faster compression (or none):
					//    - options.Snappy for decent speed, moderate compression
					//    - options.None  for fastest writes, bigger on-disk size
					//    - options.ZSTD  for better compression, more CPU cost
					// WithCompression(options.Snappy).
					// 6) Adjust number of compaction threads (fewer saves CPU/memory, might slow flush)
					WithNumCompactors(4) // default is 4; try 2 if CPU/memory is tight

				// SyncWrites is false by default, which helps write speed.
				// If you need full crash safety (OS/hardware crash), set WithSyncWrites(true) but expect slower writes.

				return opts
			},
		}
	case "sqlite":
		store = &sqlite3.SQLite3Backend{
			DatabaseURL: "file:./data/db.sqlite3?mode=rwc",
		}
	case "postgresql":
		store = &postgresql.PostgresBackend{
			DatabaseURL: "postgres://postgres:secret@localhost:5432/eventstore?sslmode=disable",
		}
	case "mem_sqlite":
		store = &sqlite3.SQLite3Backend{
			DatabaseURL: "file::memory:?mode=memory",
		}
	case "mem":
		store = &slicestore.SliceStore{
			MaxLimit: 500,
		}
	default:
		log.Fatalf("Unknown eventstore type: %s", eventstoreType)
	}

	err = store.Init()
	if err != nil {
		log.Fatalf("Failed to initialize eventstore: %v", err)
	}
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
			Limit: 500,
		},
		nostr.Filter{
			Kinds: []int{1, 5, 7},
			Limit: 500,
		},
		// existing author
		nostr.Filter{
			Authors: []string{"9e1815dfc010252a17078f9005336bbc047f551d6d7b64545052bceddecb8a2a"},
			Limit:   500,
		},
		// non existing author
		nostr.Filter{
			Authors: []string{"00e6cc3224c036b4c090d8e76262d2e9db82cd748213a78c79fc62561f175a00"},
			Limit:   500,
		},
		nostr.Filter{
			Authors: []string{
				"f81611363554b64306467234d7396ec88455707633f54738f6c4683535098cd3",
				"00e6cc3224c036b4c090d8e76262d2e9db82cd748213a78c79fc62561f175a00",
				"84e6abe4cdf74b3826e3f64b181b38e40027dcbf6d69bacb8817ac3d9fa9da37",
				"7d4a4e87f28e0e3581d4aa923494dfb5bb428abdac20db79560e2bdec853bba8",
				"c81c7999f7276387317878e59d7c321093a433977ee6811ca76dc3a9738e1869",
			},
			Limit: 500,
		},
		// existing id
		nostr.Filter{
			IDs:   []string{"fff749cdc86b18e2768fa61ce11a30f6d0ce5cd524d606cbcac82e77dbd2807b"},
			Limit: 500,
		},
		// non existing id
		nostr.Filter{
			IDs:   []string{"0000000000063995fa1a3a99584cb32fb77c10e4e929bfbd94bb40053549f100"},
			Limit: 500,
		},
		nostr.Filter{
			IDs: []string{
				"ef3f03cc0a4a2ed8224324f4f7711c67982b6e84b3f1c7919784770e09906605",
				"4c887a1c0808e1fe84f8584cd9d14e6de77190e95925cf0757086e61aa4e528b",
				"36cfbc6135deb4ca2536325b72da3d752fd8ab766fa09e4e71389fe677f8f5d6",
				"f8462dd31fcfa49df26aee0bd5528aca7bdd786ca703b55f8c20e82b1fffb0aa",
				"cefdcd1452174d166b2533d23b708637e698711f1db1fddceaba5008e47b7803",
			},
			Limit: 500,
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
