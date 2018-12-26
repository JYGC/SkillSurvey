const appConfig = require('../config.json');
const bassClass = require("./BaseService.js");

class Reporting extends bassClass.Service {
    constructor(parameters) {
        super(parameters);
    }

    Run () {
        var thisClass = this;
        var skillList = appConfig.SkillList;
        var occurenceTable = {};
        thisClass.GetOccurence(skillList, occurenceTable);
    }

    GetOccurence(skillList, occurenceTable) {
        var thisClass = this;
        if (skillList.length > 0) {
            var currentSkill = skillList.splice(0, 1);
            thisClass.dbAdapter.Word.GetOccurence({
                WordName: currentSkill,
                callback: function (rows) {
                    occurenceTable[currentSkill] = rows;
                    thisClass.GetOccurence(skillList, occurenceTable);
                }
            });
        } else {
            console.log(occurenceTable);
        }
    }
}

exports.NewService = (settings) => new Reporting(settings);