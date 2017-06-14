function placesController($scope, info, $timeout, $http) {
    $scope.app = info.app
}

placesController.resolve = {
    info: function ($q, $http, $stateParams) {
        var deferred = $q.defer();

        $http({
            method: 'GET',
            url: 'info',
            headers: {'Content-Type': 'application/json;charset=UTF-8'}
        })
            .success(function (data) {
                deferred.resolve(data)
            })
            .error(function (data) {
                deferred.reject("error value");
            });

        return deferred.promise;
    }
}