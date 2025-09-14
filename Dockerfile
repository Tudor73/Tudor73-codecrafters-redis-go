FROM golang:1.24

WORKDIR /redis
COPY go.mod go.sum ./

RUN go mod download
COPY . .

RUN go build -v -o /main ./app
EXPOSE 6379
CMD ["/main"]