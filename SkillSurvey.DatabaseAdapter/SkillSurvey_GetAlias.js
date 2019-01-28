const databaseQuery = require("./DatabaseQuery.js");

class SkillSurvey_GetAlias extends databaseQuery.DatabaseQuery {
    constructor () {
        super();

        this.dataList = [];
        this.parameterizedQuery = `SELECT
    SkillName.Name,
    SkillWordAlias.Alias
FROM
    SkillName
LEFT JOIN
    SkillWordAlias ON SkillName.Id = SkillWordAlias.SkillNameId`;
    }

}

module.exports = (parameters) => new SkillSurvey_GetAlias(parameters);