const baseClass = require("./DatabaseTable.js");

class SkillNameTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    GetAlias(parameters) {
        var getAlias = (require("./SkillName_GetAlias.js"))();

        this.runDatabaseAllCallback(getAlias.GetQuery(), getAlias.GetFlatData(), function (error, rows) {
            if (error) {
                console.log(error.message);
            } else {
                parameters.callback(rows);
            }
        });
    }
}

module.exports = (parameters) => new SkillNameTable(parameters);