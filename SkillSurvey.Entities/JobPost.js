class JobPost {
    constructor(parameters) {
        this.Id = parameters.Id;
        this.SiteId = parameters.SiteId;
        this.JobSiteNumber = parameters.JobSiteNumber
        this.Title = parameters.Title;
        this.Body = parameters.Body;
        this.DatePosted = parameters.DatePosted;
        this.City = parameters.City;
        this.Country = parameters.Country;
        this.Suburb = parameters.Suburb;
    }
}

exports.New = (parameters) => new JobPost(parameters);