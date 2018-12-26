const databaseAdapter = require("../SkillSurvey.DatabaseAdapter/DatabaseAdapter.js");
const path = require("path");
const appConfig = require('../config.json');

class Service {
    constructor(parameters) {
        this.dbAdapter = databaseAdapter({
            databaseFilePath: path.join(appConfig.AppDataFolder, appConfig.DatabaseFile)
        });
    }

    Run() {

    }
}

exports.Service = Service;