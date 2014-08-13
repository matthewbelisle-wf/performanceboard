var Controller = function(
    $http,
    $location,
    $scope
) {
    $scope.createBoard = function() {
        $http({method: 'POST', url: '/api/'}).
            success(function(data) {
                $location.path('/' + data.board);
            });
    };
};
Controller.$inject = [
    '$http',
    '$location',
    '$scope'
];

module.exports = Controller;
