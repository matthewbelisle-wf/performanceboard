var Controller = function(
    $http,
    $scope
) {
    $http({method: 'GET', url: '/api/'}).
        success(function(data) {
            $scope.boards = data.results;
        });
};
Controller.$inject = [
    '$http',
    '$scope'
];

module.exports = Controller;
