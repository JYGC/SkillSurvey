const baseClass = require("./DatabaseTable.js");

class SkillSurveyTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    GetAlias(parameters) {
        var getAlias = (require("./SkillSurvey_GetAlias.js"))();

        this.runDatabaseAllCallback(getAlias.GetQuery(), getAlias.GetFlatData(), function (error, rows) {
            if (error) {
                console.log(error.message);
            } else {
                parameters.callback(rows);
            }
        });
    }

    // GetJobCount(parameters) {
    //     var getJobCount = (require("./SkillSurvey_GetJobCount.js"))({
    //         SkillAliases: parameters.skillAliases
    //     });

    //     this.runDatabaseAllCallback(getJobCount.GetQuery(), getJobCount.GetFlatData(), function (error) {
    //         if (error) {
    //             console.log(error.message);
    //         } else {
    //             parameters.callback(rows);
    //         }
    //     });
    // }
}

module.exports = (parameters) => new SkillSurveyTable(parameters);