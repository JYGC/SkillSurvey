const databaseQuery = require("./DatabaseQuery.js");

class JobPost_GetUnProcessed extends databaseQuery.DatabaseQuery {
    constructor () {
        super();

        this.dataList = [];
        this.parameterizedQuery = `SELECT
    Id,
    SiteId,
    JobSiteNumber,
    Title,
    Body,
    City,
    Country,
    Suburb,
    CreateDate,
    PostedDate
FROM
    JobPost
WHERE
    ProcessStatus = 0`
    }
}

module.exports = (parameters) => new JobPost_GetUnProcessed(parameters);