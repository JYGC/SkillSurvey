OUTPUT_DIR=./build

mk_output_dir:
	mkdir -p ${OUTPUT_DIR}

build: mk_output_dir
	go build -o ${OUTPUT_DIR}/migrate ./cmd/migrate/main.go
