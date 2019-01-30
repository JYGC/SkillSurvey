const bassClass = require("./ReportGeneratorBaseClass.js");

class MonthlyCountReport extends bassClass.ReportGeneratorBaseClass {
    constructor(parameters) {
        super(parameters);
    }

    ProcessData () {
        this.aliasDictionary = {};
        
        // Turn flat list of names and aliases into dictionary that groups aliases based on what
        // skillname they belong to.
        var index = this.skillNamesAndAliases.length - 1;
        while (index > 0) {
            var currentSkillName = this.skillNamesAndAliases[index];

            // First alias of the skillname must be the skill name itself.
            if (!(currentSkillName.Name in this.aliasDictionary)) {
                this.aliasDictionary[currentSkillName.Name] = [currentSkillName.Name];
            }

            // Ignore {name: "name", alias: null}. Happens when skillname does have aliases.
            if (currentSkillName.Alias !== null) {
                this.aliasDictionary[currentSkillName.Name].push(currentSkillName.Alias);
            }

            index--;
        }

        this.skillNameMonthlyCount = {};
        this.GetSkillMonthlyCount({
             // avoid passing this.aliasDictionary by reference JSON stringify then JSON parsing 
            AliasDictionary: JSON.parse(JSON.stringify(this.aliasDictionary))
        });
    }

    // Creates:
    // {'ASP.NET MVC': [{ MonthYear: '2018-09', Count: 41 }, { MonthYear: '2018-10', Count: 205 }],
    // '.NET': [{ MonthYear: '2018-09', Count: 332 }, { MonthYear: '2018-10', Count: 2057 }] }
    GetSkillMonthlyCount (parameters) {
        var thisClass = this;

        // Remove currentSkillNameAlias from parameters.AliasDictionary so it isn't pass to in the
        // recursion.
        var currentSkillName = Object.keys(parameters.AliasDictionary)[0];
        var currentSkillNameAlias = parameters.AliasDictionary[currentSkillName];
        delete parameters.AliasDictionary[currentSkillName];

        thisClass.dbAdapter.JobPost.GetMonthlyCountBySkill({
            SkillNameAliases: currentSkillNameAlias,
            callback: function (rows) {
                thisClass.skillNameMonthlyCount[currentSkillName] = rows;

                if (Object.keys(parameters.AliasDictionary).length > 0) {
                    thisClass.GetSkillMonthlyCount(parameters);
                } else {
                    thisClass.returnReportCallback(thisClass.skillNameMonthlyCount);
                }
            }
        });
    }
}

exports.NewReport = (parameters) => new MonthlyCountReport(parameters);