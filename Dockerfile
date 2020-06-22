FROM golang:1.14.4 as builder

WORKDIR /go/src/app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

ARG main_path
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/server $main_path

FROM scratch

# Copy our static executable
COPY --from=builder /go/bin/server /go/bin/server
COPY --from=builder /bin/bash /bin/bash

# Run the hello binary.
ENTRYPOINT ["/go/bin/server"]