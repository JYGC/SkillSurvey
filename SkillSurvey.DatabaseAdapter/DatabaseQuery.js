class DatabaseQuery {
    constructor () {
        this.dataList = [];
    }

    GetQuery() {
        return this.parameterizedQuery;
    }

    GetFlatData() {
        var flatArray = [];

        for (var i = 0; i < this.dataList.length; i++) {
            flatArray = flatArray.concat(Object.values(this.dataList[i]));
        }

        return flatArray;
    }
}

exports.DatabaseQuery = DatabaseQuery;