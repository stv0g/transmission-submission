FROM golang:1.20-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

ADD *.go .
ADD assets/ assets/
ADD templates/ templates/

RUN go build -tags embed -o /transmission-submission .

CMD [ "/transmission-submission" ]

