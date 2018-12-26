const baseClass = require("./DatabaseTable.js");

class WordTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    AddMany(parameters) {
        var addMany = (require("./Word_AddMany.js"))({
            WordList: parameters.WordList
        });

        this.runDatabaseCallback(addMany.GetQuery(), addMany.GetFlatData(), function (error) {
            if (error) {
                console.log(error.message);
            }
        });
    }

    GetOccurence(parameters) {
        var getOccurence = (require("./Word_GetOccurence.js"))({
            WordName: parameters.WordName
        });

        this.runDatabaseAllCallback(getOccurence.GetQuery(), getOccurence.GetFlatData(), function (error, rows) {
            if (error) {
                console.log(error.message);
            } else {
                parameters.callback(rows);
            }
        });
    }
}

module.exports = (parameters) => new WordTable(parameters);