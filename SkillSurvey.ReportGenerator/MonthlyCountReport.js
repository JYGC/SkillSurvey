const bassClass = require("./ReportGeneratorBaseClass.js");

class MonthlyCountReport extends bassClass.ReportGeneratorBaseClass {
    constructor(parameters) {
        super(parameters);
    }

    ProcessData () {
        this.aliasDictionary = {};
        
        // Turn flat list of names and aliases into dictionary that groups aliases based on what
        // skillname they belong to.
        while (this.skillNamesAndAliases.length > 0) {
            var lastElement = this.skillNamesAndAliases.pop();

            // First alias of the skillname must be the skill name itself.
            if (!(lastElement.Name in this.aliasDictionary)) {
                this.aliasDictionary[lastElement.Name] = [lastElement.Name];
            }

            // Ignore {name: "name", alias: null}. Happens when skillname does have aliases.
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

        // GetFlatData() in DatabaseQuery cannot flatten paramters for prepared query if parameters
        // are flat arrays so they have to be coverted to dictionaries
        var aliases = []
        parameters.AliasDictionary[lastKey].forEach(function (item) {
            aliases.push({alias: "%" + item + "%"} /* WHERE clause is doing LIKE %?% */ );
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