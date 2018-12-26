const bassClass = require("./BaseService.js");
const textProcessor = require("../SkillSurvey.TextProcessor/TextProcessor.js");

class TextProcess extends bassClass.Service {
    constructor(parameters) {
        super(parameters);
    }

    Run () {
        var thisClass = this;
        thisClass.dbAdapter.JobPost.GetUnProcessed({
            callback: function (jobPosts) {
                thisClass.dbAdapter.ClassifiedWord.GetAlias({
                    callback: function (classifiedWords) {
                        var classifiedWordProc = new textProcessor.ClassifiedWordProcessor({
                            ClassifiedWords: classifiedWords
                        });
                        var textProc = new textProcessor.TextProcessor();
                        var wordList = null;

                        for (var i = 0; i < jobPosts.length; i++) {
                            try {
                                wordList = classifiedWordProc.GetWordList(jobPosts[i]);
                                wordList = wordList.concat(textProc.GetWordList(jobPosts[i]));

                                while (wordList.length > 0) {
                                    thisClass.dbAdapter.Word.AddMany({
                                        WordList: wordList.splice(0, 200)
                                    });
                                }

                                thisClass.dbAdapter.JobPost.SetToProcessed({
                                    JobPostId: jobPosts[i].Id
                                });
                            } catch (error) {
                                console.log(error.message);
                            }
                        }
                    }
                });
            }
        });
    }
}

exports.NewService = (settings) => new TextProcess(settings);