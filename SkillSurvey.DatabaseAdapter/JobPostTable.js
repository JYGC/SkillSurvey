const baseClass = require("./DatabaseTable.js");

class JobPosTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
        this.collection = this.database.addCollection('JobPost');
    }

    AddMany(jobPosts) {
        formatedJobPosts = [];
        jobPosts.forEach((jobPost) => {
            formatedJobPosts.push({
                SiteName: jobPost.SiteName,
                JobSiteNumber: jobPost.JobSiteNumber,
                Title: jobPost.Title,
                Body: jobPost.Body,
                PostedDate: new Date(jobPost.PostedDate).getTime(),
                City: jobPost.City,
                Country: jobPost.Country,
                Suburb: jobPost.Suburb,
                CreateDate: new Date(jobPost.CreateDate).getTime()
            });
        });
        this.collection.insert(formatedJobPosts);

        // this.insertRows((require("./JobPost_AddMany.js"))({
        //     JobPosts: parameters.JobPosts
        // }));
    }

    UpdateMany(parameters) {
        this.updateRows((require("./JobPost_UpdateMany.js"))({
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