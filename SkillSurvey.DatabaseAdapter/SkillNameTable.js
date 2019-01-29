const baseClass = require("./DatabaseTable.js");

class SkillNameTable extends baseClass.DatabaseTable {
    constructor (parameters) {
        super(parameters);
    }

    GetAlias(parameters) {
        this.getManyRows((require("./SkillName_GetAlias.js"))(), (rows) => parameters.callback(rows));
    }
}

module.exports = (parameters) => new SkillNameTable(parameters);