var Controller = function(
    $http,
    $routeParams,
    $scope
) {
    $http({method: 'GET', url: '/api/' + $routeParams.board}).
        success(function(data) {
            $scope.namespaces = data.namespaces;
        });
};
Controller.$inject = [
    '$http',
    '$routeParams',
    '$scope'
];

module.exports = Controller;
