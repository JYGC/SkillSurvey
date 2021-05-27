const bassClass = require("./SiteAdapterBaseClass.js");

class Seek extends bassClass.SiteAdapterBaseClass {
    constructor(parameters) {
        super(parameters);
        this.configSettings = require("./Seek.json");
        this.siteName = "Seek.com.au";

        this.jobPostLink = '[data-automation="jobTitle"]';

        this.titleSelector = '[data-automation="job-detail-title"]';
        this.bodySelector = '[data-automation="jobAdDetails"]';
        this.postedDateSelector = "span.FYwKg._2Bz3E.C6ZIU_4._6ufcS_4._3KSG8_4._29m7__4._2WTa0_4";
        this.citySelector = "div.FYwKg._3VxpE_4 > div:nth-child(1)";
        this.country = "Australia";
        this.suburbSelector = "div.FYwKg._3VxpE_4 > div:nth-child(2)";

        this.titleType = this.bodyType = this.postedDateType = this.cityType = this.suburbType = "text";
    }
    
    GetJobSiteNumber(url, doc) {
        return url.substring(url.lastIndexOf("/job/") + 5, url.lastIndexOf("?"));
    }

    GetPostedDate (url, doc) {
        var postedDate = new Date();

        var ageString = doc.$(this.postedDateSelector).eq(0).text().replace("Posted ", "");
        var daysOld = 0;
        var daysIndex = ageString.indexOf("d ago");

        if (daysIndex !== -1) {
            // If "h ago", the advert is 0 days old, leave daysOld as 0
            daysOld = parseInt(ageString.substring(0, daysIndex));
        }

        // setDate changes postedDate to next month if argument is more then number of days in
        // current month
        postedDate.setDate((new Date()).getDate() - daysOld);

        return postedDate;
    }
}

exports.NewAdapter = (parameters) => new Seek(parameters);