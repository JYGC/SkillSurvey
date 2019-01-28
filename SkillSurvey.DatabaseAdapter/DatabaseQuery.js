class DatabaseQuery {
    constructor () {
        this.dataList = [];
    }

    GetQuery() {
        return this.parameterizedQuery;
    }

    // converts [{o: 1, m: a, n: 01}, {m: 2, n: b, p: 02}] to [1, a, 01, 2, b, 02]. We do this as
    // node sqlite3 can only insert arguments to SQL from flat arrays
    GetFlatData() {
        var flatArray = [];

        for (var i = 0; i < this.dataList.length; i++) {
            flatArray = flatArray.concat(Object.values(this.dataList[i]));
        }

        return flatArray;
    }
}

exports.DatabaseQuery = DatabaseQuery;