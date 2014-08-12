var fs = require('fs');

var directive = function() {
    return {
        restrict: 'E',
        template: fs.readFileSync(__dirname + '/template.html', 'utf8'),
        controller: require('./controller'),
        link: function(scope, element, attrs) {
            console.log(scope.namespaces);
        }
    };
};

module.exports = directive;
