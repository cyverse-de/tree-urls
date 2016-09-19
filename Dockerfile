FROM golang:1.7-alpine

COPY . /go/src/github.com/cyverse-de/tree-urls
RUN go install -v -ldflags "-X main.appver=$version -X main.gitref=$git_commit" github.com/cyverse-de/tree-urls

EXPOSE 60000
ENTRYPOINT ["tree-urls"]
CMD ["--help"]

ARG git_commit=unknown
ARG version="2.9.0"

LABEL org.cyverse.git-ref="$git_commit"
LABEL org.cyverse.version="$version"
