# pbgen

This tool exists because I copied the
[proto description from `codetheweb/anylist`](https://github.com/codetheweb/anylist/blob/master/lib/definitions.json),
but it's in a JSON format I hadn't come across before. It seems like
[`protobufjs`](https://www.npmjs.com/package/protobufjs) is able to ingest this
file directly and then work with protos, but I didn't see any equivalent
functionality in the
[Go protobuf libraries](https://pkg.go.dev/google.golang.org/protobuf), which
makes sense because that feature relies on a bit of JS metaprogramming.

So my garbage solution was to turn that JSON file back into a `.proto`
file, which is what `pbgen` does. I don't know where the original author
(`codetheweb`) got that JSON file from, so that's the best I got.

The generated `.proto` file can then be turned into Go code with the standard
`protoc` and `protoc-gen-go` tools. For simplicity, the whole pipeline (JSON ->
Proto -> Go) can be run with `go generate` if you have the necessary tools
installed..
