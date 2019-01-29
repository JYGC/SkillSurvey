const baseClass = require("./DatabaseTable.js");

class JobPosTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    AddMany(parameters) {
        this.insertRows((require("./JobPost_AddMany.js"))({
            JobPosts: parameters.JobPosts
        }));
    }

    GetMonthlyCountBySkill(parameters) {
        this.getManyRows((require("./JobPost_GetMonthlyCountBySkill.js"))({
            SkillNameAliases: parameters.SkillNameAliases
        }), (rows) => parameters.callback(rows));
    }
}

module.exports = (parameters) => new JobPosTable(parameters);