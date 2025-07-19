OUTPUT_DIR=./build
DEPLOY_DIR=${HOME}/Testing/SkillSurvey

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

build_housekeeping:
	go build -o ${OUTPUT_DIR}/housekeeping ./cmd/housekeeping/main.go

run_housekeeping: build_housekeeping
	${OUTPUT_DIR}/housekeeping

deploy: build_survey build_reports build_results build_housekeeping
	mkdir -p ${DEPLOY_DIR}
	cp ${OUTPUT_DIR}/survey ${DEPLOY_DIR}/
	cp ${OUTPUT_DIR}/reports ${DEPLOY_DIR}/
	cp ${OUTPUT_DIR}/results ${DEPLOY_DIR}/
	cp ${OUTPUT_DIR}/*.json ${DEPLOY_DIR}/
	cp ${OUTPUT_DIR}/housekeeping ${DEPLOY_DIR}/
