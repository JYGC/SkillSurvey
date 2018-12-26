const databaseAdapter = require("../SkillSurvey.DatabaseAdapter/DatabaseAdapter.js");
const path = require("path");
const appConfig = require('../config.json');

class ServiceBaseClass {
    constructor(parameters) {
        this.dbAdapter = databaseAdapter({
            databaseFilePath: path.join(appConfig.AppDataFolder, appConfig.DatabaseFile)
        });
    }

    Run() {

    }
}

exports.ServiceBaseClass = ServiceBaseClass;