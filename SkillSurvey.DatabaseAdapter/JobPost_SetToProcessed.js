const databaseQuery = require("./DatabaseQuery.js");

class JobPost_SetToProcessed extends databaseQuery.DatabaseQuery {
    constructor (parameters) {
        super();
        var thisClass = this;
        
        // Get data to be inserted to database
        thisClass.dataList = [{
            Id: parameters.JobPostId
        }];

        thisClass.parameterizedQuery = `UPDATE JobPost
SET ProcessStatus = 1
WHERE Id = ?`
    }
}

module.exports = (parameters) => new JobPost_SetToProcessed(parameters);