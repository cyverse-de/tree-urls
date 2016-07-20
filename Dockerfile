FROM golang:1.6-alpine

ARG git_commit=unknown
LABEL org.cyverse.git-ref="$git_commit"

COPY . /go/src/github.com/cyverse-de/tree-urls
RUN go install github.com/cyverse-de/tree-urls

EXPOSE 60000
ENTRYPOINT ["tree-urls"]
CMD ["--help"]
