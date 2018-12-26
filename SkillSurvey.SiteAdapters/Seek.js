const bassClass = require("./SiteAdapter.js");

class Seek extends bassClass.SiteAdapter {
    constructor(parameters) {
        super(parameters);
        this.configSettings = require("./Seek.json");
        this.siteName = "Seek.com.au";

        this.jobPostLink = '[data-automation="jobTitle"]';
        
        var jobDescriptionTag = '[aria-labelledby="jobDescription"] ';
        var jobInfoTag = '[aria-labelledby="jobInfoHeader"] ';

        this.titleSelector = jobDescriptionTag + '[data-automation="job-detail-title"]';
        this.bodySelector = jobDescriptionTag + '[data-automation="jobDescription"] [data-automation="mobileTemplate"]';
        this.postedDateSelector = jobInfoTag + '[data-automation="job-detail-date"]';
        this.citySelector = jobInfoTag + 'dd:nth-child(4) strong';
        this.country = "Australia";
        this.suburbSelector = jobInfoTag + '> dl > dd:nth-child(4) > span > span > span';

        this.titleType = this.bodyType = this.postedDateType = this.cityType = this.suburbType = "text";
    }
    
    GetJobSiteNumber(url, doc) {
        return url.substring(url.lastIndexOf("/job/") + 5, url.lastIndexOf("?"));
    }

    GetPostedDate (url, doc) {
        return doc.$(this.postedDateSelector).eq(0).text();
    }
}

exports.NewAdapter = (parameters) => new Seek(parameters);