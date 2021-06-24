const baseClass = require("./DatabaseTable.js");

class JobPosTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    UpdateAndInsert(jobPosts) {
        this.database.loadDatabase({}, () => {
            // Covert dates from datetime to milliseconds since epoch and get SiteId for each jobPost.SiteName
            var sitesCollection = this.database.addCollection('Site');
            dbFormatJobPosts = [];
            jobPosts.forEach((jobPost) => {
                dbFormatJobPosts.push({
                    SiteId: sitesCollection.find({ Name: {$eq : jobPost.SiteName } }).$loki,
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
            
            var jobPostsCollection = this.database.addCollection('JobPost');
            
            // Update JobPosts that are already in database
            inboundJobPostSiteNumbers = []
            jobPosts.forEach((jobPost) => {
                inboundJobPostSiteNumbers.push(jobPost.SiteName);
            });
            existingJobPosts = jobPostsCollection.Find({
                JobSiteNumber: { $in: inboundJobPostSiteNumbers }
            });
            existingJobPosts.forEach((existingJobPost) => {
                var inboundIndex = -1;
                var inboundJobPost = dbFormatJobPosts.find((jobPost, index) => {
                    inboundIndex = index;
                    return jobPost.JobSiteNumber === existingJobPost.JobSiteNumber
                });
                if (!this.__IsEmptyOrSpaces(inboundJobPost.Title)) existingJobPost.Title = inboundJobPost.Title;
                if (!this.__IsEmptyOrSpaces(inboundJobPost.Body)) existingJobPost.Body = inboundJobPost.Body;
                if (!this.__IsEmptyOrSpaces(inboundJobPost.PostedDate)) existingJobPost.PostedDate = inboundJobPost.PostedDate;
                if (!this.__IsEmptyOrSpaces(inboundJobPost.City)) existingJobPost.City = inboundJobPost.City;
                if (!this.__IsEmptyOrSpaces(inboundJobPost.Country)) existingJobPost.Country = inboundJobPost.Country;
                if (!this.__IsEmptyOrSpaces(inboundJobPost.Suburb)) existingJobPost.Suburb = inboundJobPost.Suburb;
                jobPostsCollection.update(existingJobPost);
                dbFormatJobPosts.splice(inboundIndex, 1);
            });

            // Insert new JobPosts
            jobPostsCollection.insert(dbFormatJobPosts);
            this._SaveData();
        });
    }

    __IsEmptyOrSpaces(str) {
        return str === null || str.match(/^ *$/) !== null;
    }

    GetMonthlyCountBySkill(SkillNameAliases, callback) {
        this.database.loadDatabase({}, () => {
            var currentDate = new Date();
            var lastYearStartOfMth = new Date(currentDate.getFullYear() - 1, currentDate.getMonth(), 1);
            
            var jobPostsCollection = this.database.addCollection('JobPost');
            var dbFormatJobPostsWithSkillNameAliases = jobPostsCollection.find({
                $and: [{
                    $or: [
                        { Title: { $contains: SkillNameAliases } },
                        { Body: { $contains: SkillNameAliases } }
                    ]}, {
                        PostedDate: { $gte: lastYearStartOfMth.getTime() }
                    }
                ]
            });

            // Change dates from milliseconds from epoch to Date
            var jobPostsWithSkillNameAliases = [];
            dbFormatJobPostsWithSkillNameAliases.forEach((dbFormatJobPost) => {
                var jobPost = dbFormatJobPost;
                jobPost.PostedDate = new Date(dbFormatJobPost.PostedDate);
                jobPost.CreateDate = new Date(dbFormatJobPost.CreateDate);
                jobPostsWithSkillNameAliases.push(jobPost);
            });

            // Get number of JobPosts each month within the last year
            var monthlyCount = [];
            var pastYearMthDts = [Array(12).keys()].map((m) => {
                return new Date(new Date(lastYearStartOfMth).setMonth(lastYearStartOfMth.getNMonth() + m));
            });
            pastYearMthDts.forEach((mthDts) => {
                var monthStr = '' + (mthDts.getMonth() + 1);
                monthStr = (monthStr.length > 2 ? '0' : '') + monthStr;
                var count = jobPostsWithSkillNameAliases.filter((jobPosts) => {
                    return jobPosts.PostedDate.getMonth() === mthDts.getMonth()
                }).length;
                monthlyCount.push({
                    YearMonth: '' + mthDts.getFullYear() + '-' + monthStr,
                    Count: count,
                });
            });

            callback(monthlyCount);
        });
    }
}

module.exports = (parameters) => new JobPosTable(parameters);