var NAME = 'performanceboard';
module.exports = NAME;

require('angular-route'); // Just needs included once
require('bootstrap'); // Just needs included once
var angular = require('angular');
var fs = require('fs');

var app = angular.module(NAME, [
    'ngRoute',
]);

// HTML5 mode for URLs
app.config([
    '$locationProvider',
    function($locationProvider) {
        $locationProvider.html5Mode(true);
    }
]);

// Routes - Params are available by their biological taxonomy names
app.config([
    '$routeProvider',
    function($routeProvider) {
        $routeProvider.
            when('/', {template: fs.readFileSync(__dirname + '/home.html', 'utf8')}).
            when('/:board', {template: fs.readFileSync(__dirname + '/home.html', 'utf8')}).
            // when('/', {template: fs.readFileSync(__dirname + '/home.html', 'utf8')}).
            otherwise({redirectTo: '/'});
    }
]);

app.directive('pbBoards', require('./boards/directive.js'));
app.directive('pbDocs', require('./docs/directive.js'));
app.directive('pbHeader', require('./header/directive.js'));
app.directive('pbNewBoard', require('./new_board/directive.js'));
