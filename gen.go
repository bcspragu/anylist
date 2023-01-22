package main

//go:generate go run ./cmd/pbgen/ && protoc -I=pb/ --go_out=pb/ --go_opt=paths=source_relative pb/api.proto
func init() {}