OUTPUT_DIR=./build

mk_output_dir:
	mkdir -p ${OUTPUT_DIR}

build: mk_output_dir
	go build -o ${OUTPUT_DIR}/runtask ./cmd/runtask/main.go

run_dev: build
	${OUTPUT_DIR}/runtask
