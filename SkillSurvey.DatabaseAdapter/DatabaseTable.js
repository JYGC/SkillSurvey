class DatabaseTable {
    constructor (parameters) {
        this.passToExporterCallback = parameters.passToExporterCallback;
        this.runDatabaseCallback = parameters.runDatabaseCallback;
        this.runDatabaseAllCallback = parameters.runDatabaseAllCallback;
        this.runDatabaseGetCallback = parameters.runDatabaseGetCallback;
    }
}

exports.DatabaseTable = DatabaseTable;