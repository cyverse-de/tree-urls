# tree-urls

A service for the CyVerse Discovery Environment that provides CRUD access to tree-url data.

## Build

```bash
docker run --rm -v $(pwd):/go/src/github.com/cyverse-de/tree-urls -w /go/src/github.com/cyverse-de/tree-urls golang:1.6 go build -v
docker build --rm -t discoenv/tree-urls .
```
