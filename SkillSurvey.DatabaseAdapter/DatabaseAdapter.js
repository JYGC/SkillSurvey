const sqlite3 = require('sqlite3');

class DatabaseAdapter {
    constructor (parameters) {
        this.database = new sqlite3.Database(parameters.databaseFilePath);

        var thisClass = this;

        var callbackErrorHandling = function (error, rows, returnCallback) {
            if (error) {
                console.log(error.message);
            } else {
                (typeof returnCallback === "function") && returnCallback(rows);
            }
        };

        var tableParamters = {
            insertRows: function (sqlQuery, returnCallback) {
                thisClass.database.run(sqlQuery.GetQuery(), sqlQuery.GetFlatData(),
                    (error) => callbackErrorHandling(error, null, returnCallback));
            },
            getOneRow: function (sqlQuery, returnCallback) {
                thisClass.database.get(sqlQuery.GetQuery(), sqlQuery.GetFlatData(),
                    (error, rows) => callbackErrorHandling(error, rows, returnCallback));
            },
            getManyRows: function (sqlQuery, returnCallback) {
                thisClass.database.all(sqlQuery.GetQuery(), sqlQuery.GetFlatData(),
                    (error, rows) => callbackErrorHandling(error, rows, returnCallback));
            }
        };

        this.JobPost = (require('./JobPostTable.js'))(tableParamters);
        this.SkillName = (require('./SkillNameTable.js'))(tableParamters);
    }
}

module.exports = (parameters) => new DatabaseAdapter(parameters);