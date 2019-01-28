const bassClass = require("./ReportGeneratorBaseClass.js");

class MonthlyCountReport extends bassClass.ReportGeneratorBaseClass {
    constructor(parameters) {
        super(parameters);
    }

    ProcessData () {
        var tmpRawSkillNames = this.rawSkillNames;
        this.aliasDictionary = {};
        
        while (tmpRawSkillNames.length > 0) {
            var lastElement = tmpRawSkillNames.pop();

            if (!(lastElement.Name in this.aliasDictionary)) {
                this.aliasDictionary[lastElement.Name] = [lastElement.Name];
            }

            if (lastElement.Alias !== null) {
                this.aliasDictionary[lastElement.Name].push(lastElement.Alias);
            }
        }

        this.skillNameMonthlyCount = {};
        this.GetSkillMonthlyCount({
            AliasDictionaryKeys: Object.keys(this.aliasDictionary),
            AliasDictionary: this.aliasDictionary
        });
    }

    GetSkillMonthlyCount (parameters) {
        var thisClass = this;

        var lastKey = parameters.AliasDictionaryKeys.pop();

        var aliases = []

        parameters.AliasDictionary[lastKey].forEach(function (item) {
            aliases.push({alias: "%" + item + "%"});
        });

        thisClass.dbAdapter.JobPost.GetMonthlyCountBySkill({
            SkillNameAliases: aliases,
            callback: function (rows) {
                thisClass.skillNameMonthlyCount[lastKey] = rows;

                if (parameters.AliasDictionaryKeys.length > 0) {
                    thisClass.GetSkillMonthlyCount(parameters);
                } else {
                    thisClass.returnReportCallback(thisClass.skillNameMonthlyCount);
                }
            }
        });
    }
}

exports.NewReport = (parameters) => new MonthlyCountReport(parameters);