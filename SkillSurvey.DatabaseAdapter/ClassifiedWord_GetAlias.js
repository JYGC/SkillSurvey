const databaseQuery = require("./DatabaseQuery.js");

class ClassifiedWord_GetAlias extends databaseQuery.DatabaseQuery {
    constructor () {
        super();

        this.dataList = [];
        this.parameterizedQuery = `SELECT
    [Name],
    [Id]
FROM
    [ClassifiedWord]
WHERE
    [Type] = 2
UNION
SELECT
    [ClassifiedWordAlias].[Alias] AS [Name],
    [ClassifiedWordAlias].[ClassifiedWordId] AS [Id]
FROM
    [ClassifiedWordAlias]
LEFT JOIN
    [ClassifiedWord] ON [ClassifiedWordAlias].[ClassifiedWordId] = [ClassifiedWord].[Id]
WHERE
    [ClassifiedWord].[Type] = 2`;
    }
}

module.exports = (parameters) => new ClassifiedWord_GetAlias(parameters);