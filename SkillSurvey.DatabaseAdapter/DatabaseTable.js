const sqlite3 = require('sqlite3');

class DatabaseTable {
    constructor (parameters) {
        this.insertRows = parameters.insertRows;
        this.getManyRows = parameters.getManyRows;
        this.getOneRow = parameters.getOneRow;
    }
}

exports.DatabaseTable = DatabaseTable;