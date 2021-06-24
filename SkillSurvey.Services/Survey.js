const webSpider = require("node-spider");
const path = require("path");
const fs = require("fs");
const appConfig = require('../config.json');
const bassClass = require("./ServiceBaseClass.js");

class Survey extends bassClass.ServiceBaseClass {
    constructor(paramenters) {
        super(paramenters);
        var thisClass = this;

        thisClass.jobPosts = [];

        // Define web spider object
        thisClass.webSpider = new webSpider({
            concurrent: 5,
            delat: 0,
            allowDuplicates: false,
            catchErrors: true,
            addReferrer: false,
            xhr: false,
            keepAlive: false,
            error: function (error, url) {
                console.log(error);
            },
            // function that is called after web spider finishes
            done: function () {
                var totalItems = thisClass.jobPosts.length;
                var itemsInserted = 0;
                while (thisClass.jobPosts.length > 0) {
                    var jobPostsChunk = thisClass.jobPosts.splice(0, 100);
                    thisClass.dbAdapter.JobPost.UpdateAndInsert({
                        JobPosts: jobPostsChunk
                    });
                    itemsInserted += jobPostsChunk.length;
                    console.log(itemsInserted + " of " + totalItems + " job posts passed to database");
                }
            },
            headers: { 'user-agent': 'node-spider' },
            encoding: 'utf8'
        });
    }

    Run() {
        var thisClass = this;

        var adapterPath = path.join(appConfig.HomeFolder, appConfig.AdaptersFolder);
        var siteAdapter = null;

        // Get file pats containing site adapter modules and run them
        fs.readdir(adapterPath, function (error, files) { 
            if (error === null) {
                for (var i = 0; i < files.length; i++) {
                    if (files[i] != "SiteAdapterBaseClass.js" && files[i].includes(".js", files[i].length - 3)) {
                        siteAdapter = require(path.join(adapterPath, files[i])).NewAdapter({
                            // When a job post is downloaded add it to jobPosts array
                            pushJobPostCallback: function (jobDetails) {
                                thisClass.jobPosts.push({
                                    SiteName: jobDetails.SiteName,
                                    JobSiteNumber: jobDetails.JobSiteNumber,
                                    Title: jobDetails.Title,
                                    Body: jobDetails.Body,
                                    PostedDate: jobDetails.PostedDate,
                                    City: jobDetails.City,
                                    Country: jobDetails.Country,
                                    Suburb: jobDetails.Suburb
                                });
                            },
                            // Define function that passes new urls and post download actions to web spider
                            urlSubmitCallback: function (url, docProcessCallback) {
                                thisClass.webSpider.queue(url, docProcessCallback);
                            }
                        });

                        // Start web spidering and ading results to database
                        siteAdapter.FetchJobPosts();
                    }
                }
            } else {
                console.log(error);
            }
        });
    }
}

exports.NewService = (paramenters) => new Survey(paramenters);