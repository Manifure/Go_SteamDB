FROM golang:alpine
LABEL authors="manifure"
EXPOSE 8082
WORKDIR /SteamDB
COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build cmd/main.go

ENTRYPOINT ["sh", "-c", "sleep 30 && ./main"]