class Site {
    constructor (parameters) {
        this.Id = parameters.Id;
        this.Name = parameters.Name;
    }
}

exports.New = (parameters) => new Site(parameters);