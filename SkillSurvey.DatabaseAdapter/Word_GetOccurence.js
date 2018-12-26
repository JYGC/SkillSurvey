const databaseQuery = require("./DatabaseQuery.js");

class Word_GetOccurence extends databaseQuery.DatabaseQuery {
    constructor (parameters) {
        super();
        
        this.dataList = [parameters.WordName];
        this.parameterizedQuery = `SELECT
    strftime('%m-%Y', Word.JobPostCreateDate) [MonthYear],
    COUNT(Word.Id) [Occurence]
FROM
    ClassifiedWord
LEFT JOIN
    Word ON ClassifiedWord.Id = Word.ClassifiedWordId
WHERE
    ClassifiedWord.Name = ?
GROUP BY
    [MonthYear]`;
    }
}

module.exports = (parameters) => new Word_GetOccurence(parameters);