module github.com/kaijietti/graftd

go 1.17

replace (
	github.com/hashicorp/raft => ./vendor/github.com/hashicorp/raft
	github.com/hashicorp/raft-boltdb => ./vendor/github.com/hashicorp/raft-boltdb
)

require (
	github.com/hashicorp/raft v1.3.2
	github.com/hashicorp/raft-boltdb v0.0.0-20211202195631-7d34b9fb3f42
)

require (
	github.com/armon/go-metrics v0.3.8 // indirect
	github.com/boltdb/bolt v1.3.1 // indirect
	github.com/hashicorp/go-hclog v0.9.1 // indirect
	github.com/hashicorp/go-immutable-radix v1.0.0 // indirect
	github.com/hashicorp/go-msgpack v0.5.5 // indirect
	github.com/hashicorp/golang-lru v0.5.0 // indirect
	golang.org/x/sys v0.0.0-20200122134326-e047566fdf82 // indirect
)
