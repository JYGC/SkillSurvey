var XSLXChart = require("xlsx-chart");
const path = require("path");
var fs = require("fs");
const appConfig = require('../config.json');
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
    // [ { SkillName: 'ASP.NET MVC', YearMonthJobCounts: {
    //  '2019-01': 18, '2018-12': 142, '2018-11': 292, '2018-10': 205, '2018-09': 41 } },
    // { SkillName: '.NET Framework', YearMonthJobCounts: {
    //  '2019-01': 200, '2018-12': 1310, '2018-11': 2310, '2018-10': 2057, '2018-09': 332 } } ]
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
                var yearMonthJobCounts = {}

                var i = rows.length - 1;
                while (i >= 0) {
                    yearMonthJobCounts[rows[i].YearMonth] = rows[i].Count;
                    i--;
                }

                thisClass.skillNameMonthlyCount[currentSkillName] = yearMonthJobCounts;

                if (Object.keys(parameters.AliasDictionary).length > 0) {
                    thisClass.GetSkillMonthlyCount(parameters);
                } else {
                    // Create date fields from same month last year to one month ago with one month
                    // intervals ['2018-02', '2018-03', ..., '2019-01']
                    var fieldDate = new Date();
                    fieldDate.setUTCMonth(fieldDate.getUTCMonth() - 12);
                    // Set day part to 1 because when the current date day is more than 28 days and
                    // the next month has only 28 days, we get ['2018-01', '2018-03', ...]
                    fieldDate.setUTCDate(1);

                    var yearMonthFields = [];
                    var fieldMonth = null;
                    for (i = 0; i < 12; i++) {
                        fieldMonth = fieldDate.getUTCMonth() + 1;
                        yearMonthFields.push(fieldDate.getUTCFullYear().toString() +
                            ((fieldMonth < 10) ? "-0" : "-") + fieldMonth.toString());
                        fieldDate.setUTCMonth(fieldDate.getUTCMonth() + 1);
                    }

                    // Create excel report
                    var xlsxChart = new XSLXChart();
                    xlsxChart.generate({
                        chart: "line",
                        titles: Object.keys(thisClass.skillNameMonthlyCount),
                        fields: yearMonthFields,
                        data: thisClass.skillNameMonthlyCount,
                        chartTitle: "Monthly Count Report"
                    }, function (err, data) {
                        if (err) {
                            console.error(err);
                        } else {
                            if (!fs.existsSync(appConfig.ReportFolder)){
                                fs.mkdirSync(appConfig.ReportFolder);
                            }
                            fs.writeFileSync(path.join(appConfig.ReportFolder, "MonthlyCountReport " +
                                (new Date()).toDateString() + ".xlsx"), data);
                        }
                    });
                }
            }
        });
    }
}

exports.NewReport = (parameters) => new MonthlyCountReport(parameters);
