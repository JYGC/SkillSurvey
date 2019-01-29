class ReportGeneratorBaseClass {
    constructor (parameters) {
        this.returnReportCallback = parameters.returnReportCallback; // Get function that returns report object
        this.dbAdapter = parameters.dbAdapter;
    }

    // From outside, call this only
    GetReport (parameters) {
        this.GetSkillNamesAndAliases(parameters);
    }

    GetSkillNamesAndAliases (parameters) {
        var thisClass = this;
        thisClass.dbAdapter.SkillName.GetAlias({
            callback: function (rows) {
                thisClass.skillNamesAndAliases = rows; // set to class property to work with asynchronous callbacks
                thisClass.ProcessData();
            }
        });
    }

    ProcessData () {
    }
}

exports.ReportGeneratorBaseClass = ReportGeneratorBaseClass;