const databaseQuery = require("./DatabaseQuery.js");

class JobPost_GetMonthlyCountBySkill extends databaseQuery.DatabaseQuery {
    constructor (parameters) {
        super();
        var thisClass = this;
        
        // Get data to be inserted to database
        thisClass.dataList = [];
        parameters.SkillNameAliases.forEach(function (item) {
            thisClass.dataList.push("%" + item + " %"); /* WHERE clause is doing LIKE %word % */
            thisClass.dataList.push("%" + item + ",%"); /* WHERE clause is doing LIKE %word.% */
            thisClass.dataList.push("%" + item + ".%"); /* WHERE clause is doing LIKE %word,% */
            thisClass.dataList.push("%" + item + "\n%"); /* WHERE clause is doing LIKE %word\n% */
        });

        thisClass.parameterizedQuery = `SELECT
    strftime('%Y-%m', JobPost.PostedDate) [YearMonth],
    COUNT(JobPost.Id) [Count]
FROM
    JobPost
WHERE`
        thisClass.parameterizedQuery += `
    JobPost.Body LIKE ?
OR`.repeat(thisClass.dataList.length) // Number of where clauses must equal number of parameters
    .replace(/OR$/,`GROUP BY
    [YearMonth];`); // Remove last OR replace with GROUP BY clause
;
    }
}

module.exports = (parameters) => new JobPost_GetMonthlyCountBySkill(parameters);