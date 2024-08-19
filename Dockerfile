# SPDX-FileCopyrightText: 2023 Steffen Vogel <post@steffenvogel.de>
# SPDX-License-Identifier: Apache-2.0

FROM golang:1.23-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

ADD *.go .
ADD assets/ assets/
ADD templates/ templates/

RUN go build -tags embed -o /transmission-submission .

CMD [ "/transmission-submission" ]

