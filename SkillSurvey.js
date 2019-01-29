var serviceName = null;

var arguments = process.argv;
for (var i = 0; i < arguments.length; i++) {
    var argValueArray = arguments[i].split("=");

    switch (argValueArray[0]) {
        case "service":
        serviceName = argValueArray[1];
            break;
    }
}

var serviceImport = null
switch (serviceName) {
    case "survey":
        serviceImport = require('./SkillSurvey.Services/Survey.js');
        break;
    case "reports":
        serviceImport = require('./SkillSurvey.Services/Reporting.js');
        break;
}

var service = serviceImport.NewService(null);
service.Run();