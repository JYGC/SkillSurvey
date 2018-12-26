const databaseQuery = require("./DatabaseQuery.js");

class Word_AddMany extends databaseQuery.DatabaseQuery {
    constructor(parameters) {
        super();
        var thisClass = this;

        // Get data to be inserted to database
        thisClass.dataList = [];
        parameters.WordList.forEach(function (wordList) {
            thisClass.dataList.push({
                Name: wordList.Name,
                JobPostCreateDate: wordList.JobPostCreateDate
            });
        });

        thisClass.parameterizedQuery = `INSERT INTO [Word]
(
    [Name],
    [ClassifiedWordId],
    [JobPostCreateDate]
)
SELECT
    InputData.Name,
    ClassifiedWords.Id,
    InputData.JobPostCreateDate
FROM (`;
        thisClass.parameterizedQuery += `
    SELECT
        ? Name,
        ? JobPostCreateDate
    UNION`.repeat(parameters.WordList.length)
            .replace(/UNION$/, ") InputData");
        thisClass.parameterizedQuery += `
LEFT JOIN
(
    SELECT
        [ClassifiedWord].[Id],
        [Name],
        [Type]
    FROM
        [ClassifiedWord]
    UNION
    SELECT
        [ClassifiedWordAlias].[ClassifiedWordId] AS [Id],
        [ClassifiedWordAlias].[Alias] AS [Name],
        [ClassifiedWord].[Type]
    FROM
        [ClassifiedWordAlias]
    LEFT JOIN
        [ClassifiedWord] ON [ClassifiedWordAlias].[ClassifiedWordId] = [ClassifiedWord].[Id]
) ClassifiedWords ON InputData.Name LIKE ClassifiedWords.Name --CASE INSENSTIVE COMPARE
WHERE
    IFNULL(ClassifiedWords.Type, 0) != 1`
    }
}

module.exports = (parameters) => new Word_AddMany(parameters);