OUTPUT_DIR=./build

build_config:
	cp *.json ${OUTPUT_DIR}/

build_survey:
	go build -o ${OUTPUT_DIR}/survey ./cmd/survey/main.go

run_survey: build_survey
	${OUTPUT_DIR}/survey
