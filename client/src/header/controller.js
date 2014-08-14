var Controller = function(
    $http,
    $routeParams,
    $scope
) {
    var boardApi = '/api/' + $routeParams.board;
    $http({method: 'GET', url: boardApi}).
        success(function(data) {
            $scope.boardApi = boardApi;
        });
};
Controller.$inject = [
    '$http',
    '$routeParams',
    '$scope'
];

module.exports = Controller;
