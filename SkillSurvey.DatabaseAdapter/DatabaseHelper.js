class DatabaseHelper {
    static MakeInsertPlaceholder(parameters) {
        var placeholderRow = "(" + parameters.columnList.map((column) => "?").join(",") + ")";
        var valuesPlaceholders = parameters.inputEntryList.map((entry) => placeholderRow).join(",\n");

        return valuesPlaceholders;
    }
}

module.exports = DatabaseHelper;