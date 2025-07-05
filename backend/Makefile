OUTPUT_DIR=./build

mk_output_dir:
	mkdir -p ${OUTPUT_DIR}

cp_config: mk_output_dir
	cp *.json ${OUTPUT_DIR}/

build_survey: cp_config
	go build -o ${OUTPUT_DIR}/survey ./cmd/survey/main.go

run_survey: build_survey
	${OUTPUT_DIR}/survey

build_reports: cp_config
	go build -o ${OUTPUT_DIR}/reports ./cmd/reports/main.go

run_reports: build_reports
	${OUTPUT_DIR}/reports

build_results: cp_config
	go build -o ${OUTPUT_DIR}/results ./cmd/results/main.go

run_results: build_results
	${OUTPUT_DIR}/results
