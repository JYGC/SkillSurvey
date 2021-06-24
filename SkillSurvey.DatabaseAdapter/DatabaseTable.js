const sqlite3 = require('sqlite3');

class DatabaseTable {
    constructor (parameters) {
        this.database = parameters.database;
        this.getErrorCallback = parameters.getErrorCallback;
    }

    _SaveData() {
        this.database.saveDatabase((error) => {
            if (error && typeof(this.getErrorCallback) === 'function') {
                this.getErrorCallback(error);
            }
        });
    }
}

exports.DatabaseTable = DatabaseTable;