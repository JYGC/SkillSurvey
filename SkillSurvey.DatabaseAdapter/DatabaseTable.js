const sqlite3 = require('sqlite3');

class DatabaseTable {
    constructor (parameters) {
        this.database = new (require('lokijs'))(parameters.databaseFilePath);



        
        this.insertRows = parameters.insertRows;
        this.updateRows = parameters.updateRows;
        this.getManyRows = parameters.getManyRows;
        this.getOneRow = parameters.getOneRow;
    }
}

exports.DatabaseTable = DatabaseTable;