const baseClass = require("./DatabaseTable.js");

class ClassifiedWordTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    GetAlias(parameters) {
        var getAlias = (require("./ClassifiedWord_GetAlias.js"))();

        this.runDatabaseAllCallback(getAlias.GetQuery(), getAlias.GetFlatData(), function (error, rows) {
            if (error) {
                console.log(error.message);
            } else {
                parameters.callback(rows);
            }
        });
    }
}

module.exports = (parameters) => new ClassifiedWordTable(parameters);