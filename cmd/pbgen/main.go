package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"text/template"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

type Payload struct {
	Package  string    `json:"package"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Name   string  `json:"name"`
	Fields []Field `json:"fields"`
	Enums  []Enum  `json:"enums"`
}

type Field struct {
	Rule string `json:"rule"`
	Type string `json:"type"`
	Name string `json:"name"`
	ID   int    `json:"id"`
}

type Enum struct {
	Name   string      `json:"name"`
	Values []EnumValue `json:"values"`
}

type EnumValue struct {
	Name string `json:"name"`
	ID   int    `json:"id"`
}

func run() error {
	f, err := os.Open("pb/description.json")
	if err != nil {
		return fmt.Errorf("failed to open pb desc file: %w", err)
	}
	defer f.Close()

	var pl Payload
	if err := json.NewDecoder(f).Decode(&pl); err != nil {
		return fmt.Errorf("failed to decode pb desc to JSON: %w", err)
	}

	tmpl, err := template.New("proto.tmpl").Funcs(map[string]any{
		"maybeRule": func(f Field) string {
			if f.Rule == "repeated" {
				return "repeated "
			}
			return ""
		},
	}).ParseFiles("cmd/pbgen/proto.tmpl")
	if err != nil {
		return fmt.Errorf("failed to init proto template: %w", err)
	}

	out, err := os.Create("pb/api.proto")
	if err != nil {
		return fmt.Errorf("failed to open out file: %w", err)
	}
	defer out.Close()

	if err := tmpl.Execute(out, pl); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	if err := out.Close(); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
