FROM golang:1.26-alpine3.23

WORKDIR /app

COPY . .

RUN go mod tidy
RUN go build -o app .

RUN addgroup --system appgroup \
    && adduser --system appuser --ingroup appgroup

USER appuser

CMD ["./app"]
