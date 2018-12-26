const databaseQuery = require("./DatabaseQuery.js");

// Add multiple job posts to JobPost table
class JobPost_AddMany extends databaseQuery.DatabaseQuery {
    constructor (parameters) {
        super();
        var thisClass = this;

        // Get data to be inserted to database
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

        thisClass.parameterizedQuery = `INSERT INTO [JobPost]
(
    SiteId,
    JobSiteNumber,
    Title,
    Body,
    PostedDate,
    City,
    Country,
    Suburb
)
SELECT
    [Site].Id,
    InputData.JobSiteNumber,
    InputData.Title,
    InputData.Body,
    InputData.PostedDate,
    InputData.City,
    InputData.Country,
    InputData.Suburb
FROM (`;
        thisClass.parameterizedQuery += `
    SELECT
        ? SiteName,
        ? JobSiteNumber,
        ? Title,
        ? Body,
        ? PostedDate,
        ? City,
        ? Country,
        ? Suburb
    UNION`.repeat(parameters.JobPosts.length) // Number of SELECTs must eqaul number of JobPosts to be entered into database
            .replace(/UNION$/, ") InputData"); // Replace UNION at end with ) InputData 
        thisClass.parameterizedQuery += `
LEFT JOIN
    [Site] ON InputData.SiteName = [Site].Name
LEFT JOIN
    [JobPost] ON InputData.JobSiteNumber = [JobPost].JobSiteNumber AND [Site].Id = [JobPost].SiteId
WHERE
    [JobPost].JobSiteNumber IS NULL;`;
    }
}

module.exports = (parameters) => new JobPost_AddMany(parameters);