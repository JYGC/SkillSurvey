const baseClass = require("./DatabaseTable.js");
const entities = require("../SkillSurvey.Entities/EntitiesHelper.js");

class SiteTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    GetByName(parameters) {
        var site = null;
        var siteName = parameters.SiteName;

        this.passToExporterCallback({
            table: 'JobPost',
            where: 'Name = ' + siteName
        }, function (error, json) {
            site = entities.NewSite(json[0]);
        });

        return site;
    }
}

module.exports = (parameters) => new SiteTable(parameters);