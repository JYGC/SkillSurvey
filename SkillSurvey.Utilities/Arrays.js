class Arrays {
    constructor () { }

    static RemoveByValues (targetArray, removeList) {
        var newArray = targetArray;
        for (var i = 0; i < removeList.length; i++) {
            newArray = newArray.filter(v => v !== removeList[i]);
        }

        return newArray;
    } 
}

module.exports = Arrays;