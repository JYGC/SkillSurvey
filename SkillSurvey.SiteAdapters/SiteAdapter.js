const webSpider = require("node-spider");

class SiteAdapter {
    constructor(parameters) {
        this.pushJobPostCallback = parameters.pushJobPostCallback;
        this.urlSubmitCallback = parameters.urlSubmitCallback;
    }

    FetchJobPosts() {
        var thisClass = this;
        var searchCriteria = null;
        var p = null;
        var fullUrl = null;
        for (var i = 0; i < this.configSettings.SearchCriteria.length; i++) {
            searchCriteria = this.configSettings.SearchCriteria[i];
            p = 1;
            while (p <= this.configSettings.Pages) {
                fullUrl = searchCriteria.Url.replace(this.configSettings.PageFlag, p.toString());
                console.log("Spidering: " + fullUrl);
                this.urlSubmitCallback(fullUrl, function (doc){
                    doc.$(thisClass.jobPostLink).each(function (i, elem) {
                        var href = doc.$(elem).attr('href');
                        var url = doc.resolve(href);
                        thisClass.FetchJobPost(url);
                    });
                });

                p++;
            }
        }
    }

    FetchJobPost (url) {
        var thisClass = this;
        this.urlSubmitCallback(url, function (doc) {
            try {
                thisClass.pushJobPostCallback({
                    SiteName: thisClass.siteName,
                    JobSiteNumber: thisClass.GetJobSiteNumber(url, doc),
                    Title: thisClass.ExtractText(doc.$(thisClass.titleSelector).eq(0), thisClass.titleType),
                    Body: thisClass.ExtractText(doc.$(thisClass.bodySelector).eq(0), thisClass.bodyType),
                    PostedDate: thisClass.GetPostedDate(url, doc),
                    City: thisClass.ExtractText(doc.$(thisClass.citySelector).eq(0), thisClass.cityType),
                    Country: thisClass.country,
                    Suburb: thisClass.ExtractText(doc.$(thisClass.suburbSelector).eq(0), thisClass.suburbType)
                });
            } catch (error) {
                console.log(error);
            }
        });
    }

    GetJobSiteNumber (url, doc) {

    }

    GetPostedDate (url, doc) {

    }

    ExtractText (htmlControl, type) {
        var text = null;

        if (type === 'text') {
            text = htmlControl.text();
        } else if (type === 'value') {
            text = htmlControl.val();
        }

        return text;
    }
}

exports.SiteAdapter = SiteAdapter;