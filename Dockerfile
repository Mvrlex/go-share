
# Build the application

FROM golang:1.23 as build

WORKDIR /go/src/go-share
COPY . .

RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /go/bin/go-share

# Create the distroless image with a non-root user

FROM gcr.io/distroless/static-debian12
COPY --from=build /go/bin/go-share /home/nonroot

ENV GOSHARE_HOST=http://localhost:8080/
ENV GOSHARE_MAX_FILE_SIZE=104857600
ENV GOSHARE_DISK_SPACE=32212254720
ENV GOSHARE_PORT=8080

USER nonroot:nonroot
EXPOSE $GOSHARE_PORT

WORKDIR /home/nonroot
CMD ["/home/nonroot/go-share"]
