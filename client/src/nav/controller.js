var Controller = function($routeParams, $scope) {
    $scope.tabs = [
        {
            name: 'All',
            href: '/' + $routeParams.board,
            active: !$routeParams.binType
        },
        {
            name: 'Day',
            href: '/' + $routeParams.board + '/day',
            active: $routeParams.binType == 'day'
        },
        {
            name: 'Hour',
            href: '/' + $routeParams.board + '/hour',
            active: $routeParams.binType == 'hour'
        },
        {
            name: 'Minute',
            href: '/' + $routeParams.board + '/minute',
            active: $routeParams.binType == 'minute'
        },
        {
            name: 'Second',
            href: '/' + $routeParams.board + '/second',
            active: $routeParams.binType == 'second'
        }
    ];
};
Controller.$inject = [
    '$routeParams',
    '$scope'
];

module.exports = Controller;
