class TextProcessor {
    constructor() {
    }

    GetWordList (jobPost) {
        var wordList = [];
        var jobPostTextArray = jobPost.Body.split(/\s+/);
        var currentWord = null;
        var index = null;

        while (jobPostTextArray.length > 0) {
            currentWord = jobPostTextArray.splice(0, 1)
            wordList.push(this.ConvertWordObject(currentWord[0], jobPost));
            index = jobPostTextArray.indexOf(currentWord);
            while (index > -1) {
                jobPostTextArray.splice(index, 1);
                index = jobPostTextArray.indexOf(currentWord);
            }
        }

        return wordList;
    }

    ConvertWordObject(currentWordString, jobPost) {
        return {
            Name: currentWordString.replace(/[.,]$/, ""),
            JobPostCreateDate: jobPost.PostedDate
        };
    }
}

class ClassifiedWordProcessor extends TextProcessor {
    constructor(parameters) {
        super();

        // Sort by longest name first
        var sortedClassifiedWords = parameters.ClassifiedWords.sort(function (a, b) {
            return b.Name.length - a.Name.length;
        });

        // Group aliases by recognized classified words
        this.classifiedWordGroups = {};
        for (var i = 0; i < sortedClassifiedWords.length; i++) {
            if (sortedClassifiedWords[i].Id in this.classifiedWordGroups) {
                this.classifiedWordGroups[sortedClassifiedWords[i].Id].push(sortedClassifiedWords[i].Name);
            } else {
                this.classifiedWordGroups[sortedClassifiedWords[i].Id] = [sortedClassifiedWords[i].Name];
            }
        }
    }

    GetWordList (jobPost) {
        var wordList = [];

        var regEx = null;
        for (var classifiedWordGroupId in this.classifiedWordGroups) {
            var classifiedWordGroup = this.classifiedWordGroups[classifiedWordGroupId]
            
            regEx = new RegExp(classifiedWordGroup.join("|"), "gi");

            if (regEx.test(jobPost.Body)) {
                wordList.push(this.ConvertWordObject(classifiedWordGroup[0], jobPost));
                jobPost.Body = jobPost.Body.replace(regEx, " ");
            }
        }

        return wordList;
    }
}

exports.TextProcessor = TextProcessor;
exports.ClassifiedWordProcessor = ClassifiedWordProcessor;