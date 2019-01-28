const appConfig = require('../config.json');
const bassClass = require("./ServiceBaseClass.js");
const monthlyCountReport = require("../SkillSurvey.ReportGenerator/MonthlyCountReport.js");

class Reporting extends bassClass.ServiceBaseClass {
    constructor(parameters) {
        super(parameters);
    }

    Run () {
        var monthlyCountReportObject = monthlyCountReport.NewReport({
            returnReportCallback: function (outputReport) {
                console.log(outputReport);
            },
            dbAdapter: this.dbAdapter
        });

        monthlyCountReportObject.GetReport();
    }
}

exports.NewService = (settings) => new Reporting(settings);