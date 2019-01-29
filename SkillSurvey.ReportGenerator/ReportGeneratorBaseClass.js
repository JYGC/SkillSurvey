class ReportGeneratorBaseClass {
    constructor (parameters) {
        this.returnReportCallback = parameters.returnReportCallback; // Get function that returns report object
        this.dbAdapter = parameters.dbAdapter;
    }

    GetReport (parameters) {
        this.GetSkillNames(parameters);
    }

    GetSkillNames (parameters) {
        var thisClass = this;
        thisClass.dbAdapter.SkillName.GetAlias({
            callback: function (rows) {
                thisClass.rawSkillNames = rows;
                thisClass.ProcessData();
            }
        });
    }

    ProcessData () {
    }
}

// class SkillNameReportBaseClass extends ReportGeneratorBaseClass {
//     constructor (parameters) {
//         super(parameters);
//     }
// }

exports.ReportGeneratorBaseClass = ReportGeneratorBaseClass;