const bassClass = require("./SiteAdapterBaseClass.js");

class Jora extends bassClass.SiteAdapterBaseClass {
    constructor(parameters) {
        super(parameters);
        this.configSettings = require("./Jora.json");
        this.siteName = "au.Jora.com";

        this.jobPostLink = '.job-item';

        this.titleSelector = 'h3.job-title.heading-xxlarge';
        this.bodySelector = '#job-description-container';
        // this.postedDateSelector not used here
        this.citySelector = '.location';
        this.country = "Australia";
        this.suburbSelector = '.location';

        this.titleType = this.bodyType = this.postedDateType = this.cityType = this.suburbType = "text";
    }
    
    GetJobSiteNumber(url, doc) {
        return url.substring(url.lastIndexOf("/job/") + 5, url.lastIndexOf("?"));
    }

    // Advertisement's post date can calculated by subtracting how old the advert is in days from
    // the current date.
    GetPostedDate (url, doc) {
        var ageString = doc.$('.date').text(); // .date contains either "N days" or "today"
        var daysOld = 0;
        var daysIndex = ageString.indexOf("days");
        var currentDate = new Date();
        var postedDate = new Date();

        if (daysIndex !== -1) {
            // If "today", the advert is 0 days old, leave daysOld as 0
            daysOld = parseInt(ageString.substring(0, daysIndex - 1));
        }

        // setDate changes postedDate to next month if argument is more then number of days in
        // current month
        postedDate.setDate(currentDate.getDate() - daysOld);
        
        return postedDate.toISOString();
    }
}

exports.NewAdapter = (parameters) => new Jora(parameters);