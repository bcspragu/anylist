FROM golang as build

WORKDIR /project

COPY go.mod /project
COPY go.sum /project
COPY main.go /project/main.go
COPY anylist/ /project/anylist
COPY pb/ /project/pb

RUN GOOS=linux CGO_ENABLED=0 go build -o server .

FROM gcr.io/distroless/static-debian11
COPY --from=build /project/server /
COPY secrets.enc.json /
CMD ["/server"]
