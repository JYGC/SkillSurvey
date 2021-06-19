const databaseQuery = require("./DatabaseQuery.js");

class JobPost_UpdateMany extends databaseQuery.DatabaseQuery {
    constructor (parameters) {
        super();
        var thisClass = this;

        thisClass.dataList = [];
        parameters.JobPosts.forEach(function (jobPost) {
            thisClass.dataList.push({
                SiteName: jobPost.SiteName,
                JobSiteNumber: jobPost.JobSiteNumber,
                Title: jobPost.Title,
                Body: jobPost.Body,
                PostedDate: (new Date(jobPost.PostedDate)).toISOString().replace(/T/, " ").replace(/Z/, ""),
                City: jobPost.City,
                Country: jobPost.Country,
                Suburb: jobPost.Suburb
            });
        });
        thisClass.parameterizedQuery = ``;
        thisClass.parameterizedQuery += `UPDATE [JobPost]
SET Title = ?,
Body = ?,
PostedDate = ?,
City = ?,
Country = ?,
Suburb = ?
WHERE SiteName = ? AND JobSiteNumber = ?;\n`.repeat(parameters.JobPosts.length);
    }
}

module.exports = (parameters) => new JobPost_UpdateMany(parameters);