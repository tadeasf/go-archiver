module go-archiver

go 1.21

require (
	github.com/go-python/gopy v0.4.10
)

replace (
	go-archiver/archiver => ./src/go_archiver/go/archiver
	go-archiver/bindings => ./src/go_archiver/go/bindings
)
