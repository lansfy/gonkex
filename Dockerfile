FROM golang:1.14 as build

ENV GOOS linux
ENV GOARCH amd64
ENV CGO_ENABLED 0

WORKDIR /build

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

# This is the ‘magic’ step that will download all the dependencies that are specified in
# the go.mod and go.sum file.
# Because of how the layer caching system works in Docker, the go mod download
# command will _only_ be re-run when the go.mod or go.sum file change
RUN go mod download

COPY . .
RUN make build

FROM alpine:3.10
LABEL Author="Lansfy <5764541+lansfy@users.noreply.github.com>"

COPY --from=build /build/gonkex /bin/gonkex
ENTRYPOINT ["/bin/gonkex"]
CMD ["-spec=/gonkex/swagger.yaml", "-host=${HOST_ARG}"]
