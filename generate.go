//go:generate protoc --go_out=. --go_opt=paths=source_relative ./api/common.proto
//go:generate protoc --plugin=/usr/bin/protoc-gen-ts_proto --ts_proto_out=./gui/src --ts_proto_opt=esModuleInterop=true ./api/common.proto
//go:generate esbuild ./gui/src/main.ts --bundle --outfile=./gui/dist/gui.js --sourcemap --minify

package main
