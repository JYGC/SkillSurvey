const sqlite3 = require('sqlite3');
const sqliteJson = require('sqlite-json');

class DatabaseAdapter {
    constructor (parameters) {
        this.database = new sqlite3.Database(parameters.databaseFilePath);
        this.jsonExporter = sqliteJson(this.database);

        var thisClass = this;
        var tableParamters = {
            runDatabaseCallback: function (parameterizedSql, sqlParameters, errorCallback) {
                thisClass.database.run(parameterizedSql, sqlParameters, errorCallback);
            },
            runDatabaseGetCallback: function (parameterizedSql, sqlParameters, returnCallback) {
                thisClass.database.get(parameterizedSql, sqlParameters, returnCallback);
            },
            runDatabaseAllCallback: function (parameterizedSql, sqlParameters, returnCallback) {
                thisClass.database.all(parameterizedSql, sqlParameters, returnCallback);
            },
            passToExporterCallback: function (querySections, returnCallback) {
                thisClass.jsonExporter.json(querySections, returnCallback);
            }
        };

        this.JobPost = (require('./JobPostTable.js'))(tableParamters);
        this.SkillName = (require('./SkillNameTable.js'))(tableParamters);
    }
}

module.exports = (parameters) => new DatabaseAdapter(parameters);