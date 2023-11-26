FROM golang:1.21

WORKDIR /app

COPY . .

RUN go build -o /bin/agent /app/cmd/agent
RUN go build -o /bin/server /app/cmd/server
