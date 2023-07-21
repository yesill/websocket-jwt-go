FROM golang:latest

WORKDIR /build

COPY . .
RUN go build -o main .

EXPOSE 8000

ENTRYPOINT [ "/build/main" ]
