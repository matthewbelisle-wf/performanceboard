var fs = require('fs');

var directive = function() {
    return {
        restrict: 'E',
        template: fs.readFileSync(__dirname + '/template.html', 'utf8')
    };
};

module.exports = directive;
