## Benchmark eventstores

This software was created to compare memory only eventstore (based on fiatjaf's, slicestore) to other eventsotres provided bi gtihub.com/fiatjaf/eventstore.

params are:

go run . <store_type> [load]

where "store_type" is one of mem, lmdb, badger, sqlite

load indicates you do not have a database ready and will need to load it from events.json file (always true for "mem").

you can get an working events.json from https://girino.org/nostr/events.json.zst

(it's just a list of events in a json file)

this table describes the results i've got up to now:

| **Load Time**           | slicestore: 885.7112 ms | lmdb: 16m 9.6313045s | badger: 27.308578 s |  |  |
|-------------------------|------------------------|----------------------|---------------------|--------------------------------------------------------|----------------------------------------------------------|
| **Filters**             | slicestore Average Time (µs) | lmdb Average Time (µs) | badger Average Time (µs) | Relative Performance Increase (lmdb to slicestore) (%) | Relative Performance Increase (badger to slicestore) (%) |
| {"limit":500}           | 273.029               | 2432.304          | 5218.979          | 790.9                                                  | 1811.4                                                   |
| {"kinds":[1],"limit":500} | 373.624               | 1304.388          | 4345.981          | 249.1                                                  | 1063.1                                                   |
| {"kinds":[1,5,7],"limit":500} | 313.219               | 954.542           | 4591.242          | 204.7                                                  | 1365.8                                                   |
| 1 author                | 329.065               | 1332.927          | 4419.829          | 305.0                                                  | 1243.5                                                   |
| 1 author (nonexistent)  | 1.032                 | 9.217             | 26.088            | 793.1                                                  | 2430.2                                                   |
| 5 authors               | 426.277               | 1402.495          | 8984.9            | 229.0                                                  | 2007.7                                                   |
| 1 id                    | 3.019                 | 13.672            | 47.848            | 352.8                                                  | 1484.6                                                   |
| 1 id (nonexistent)      | 2.095                 | 8.616             | 27.303            | 311.2                                                  | 1203.5                                                   |
| 5 ids                   | 7.613                 | 55.666            | 139.718           | 631.1                                                  | 1734.4                                                   |
