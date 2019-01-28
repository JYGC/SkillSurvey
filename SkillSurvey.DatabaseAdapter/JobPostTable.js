const baseClass = require("./DatabaseTable.js");

class JobPosTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    AddMany(parameters) {
        var addMany = (require("./JobPost_AddMany.js"))({
            JobPosts: parameters.JobPosts
        });

        this.runDatabaseCallback(addMany.GetQuery(), addMany.GetFlatData(), function (error) {
            if (error) {
                console.log(error.message);
            }
        });
    }

    GetUnProcessed(parameters) {
        var getUnProcessed = (require("./JobPost_GetUnProcessed.js"))();

        this.runDatabaseAllCallback(getUnProcessed.GetQuery(), getUnProcessed.GetFlatData(), function (error, rows) {
            if (error) {
                console.log(error.message);
            } else {
                parameters.callback(rows);
            }
        });
    }

    SetToProcessed(parameters) {
        var setToProcessed = (require("./JobPost_SetToProcessed.js"))({
            JobPostId: parameters.JobPostId
        });

        this.runDatabaseCallback(setToProcessed.GetQuery(), setToProcessed.GetFlatData(), function (error) {
            if (error) {
                console.log(error.message);
            }
        });
    }

    GetMonthlyCountBySkill(parameters) {
        var getMonthlyCountBySkill = (require("./JobPost_GetMonthlyCountBySkill.js"))({
            SkillNameAliases: parameters.SkillNameAliases
        });

        this.runDatabaseAllCallback(getMonthlyCountBySkill.GetQuery(), getMonthlyCountBySkill.GetFlatData(), function (error, rows) {
            if (error) {
                console.log(error.message);
            } else {
                parameters.callback(rows);
            }
        });
    }
}

module.exports = (parameters) => new JobPosTable(parameters);