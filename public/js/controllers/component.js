function placeController($scope, info, $timeout, $http, $stateParams, $location, $anchorScroll) {
    $scope.root = info
    $scope.selected = null
    $scope.rootComponent = $stateParams.component
    $scope.connection = {}
    $scope.shaked = null
    $scope.selectedTab = 'info'
    $scope.setSelected = function (s) {
        $scope.selectedTab = 'info'
        if ($scope.selected == s) {
            $scope.selected = null
        } else {
            $scope.selected = s
            $location.search('select', window.encodeURIComponent(s.url));
        }
    }
    $scope.setSelectedTab = function (s) {
        $scope.selectedTab = s
    }
    $scope.$on('$destroy', function () {
        $scope.isDestroed = true
    });
    $scope.update = function (info) {
        $scope.info = $scope.info || {app: {components: []}}
        if (!info.app) return
        if (!info.app.components) return

        $scope.connection.lastUpdated = info.lastUpdated

        for (var i = 0; i < info.app.components.length; i++) {
            if (info.app.components[i].app.name == $stateParams.component) {
                $scope.info.lastUpdated = info.app.components[i].lastUpdated
                $scope.info.error = info.app.components[i].error
                // components
                var components = info.app.components[i].app.components
                components = components || []
                for (var j = 0; j < components.length; j++) {
                    var c = $scope.find($scope.info.app.components, components[j].url)
                    if (!c) {
                        components[j].original = $scope.toJson(components[j].original)
                        $scope.info.app.components.push(components[j])
                    } else {
                        c.app = components[j].app
                        c.lastUpdated = components[j].lastUpdated
                        c.error = components[j].error
                        c.original = $scope.toJson(components[j].original)
                    }
                }
                break
            }
        }
    }
    $scope.toJson= function (text) {
        try {
            return JSON.parse(text)
        } catch (e) {
            return text
        }
    }
    $scope.find = function (list, url) {
        for (var i = 0; i < list.length; i++) {
            if (list[i].url == url) {
                return list[i]
                break
            }
        }
        return null
    }
    $scope.update(info)
    $scope.listen = function (delay) {
        $timeout(function () {
            $http({
                method: 'GET',
                url: "/info",
                headers: {'Content-Type': 'application/json;charset=UTF-8'}
            })
                .success(function (info) {
                    $scope.connection.error = null
                    $scope.update(info)
                })
                .error(function (error, error2, error3) {
                    $scope.connection.error = "Connection with server is lost"
                });
            if (!!$scope.isDestroed) return
            $scope.listen(5000)
        }, delay, true);
    }
    $scope.listen(5000)

    /***
     *
     */
    $scope.goToPlaces = function () {
        $location.path('#')
    }
    //Select who must be selected
    if ($location.search().select) {
        if (!$scope.info.app) return
        if (!$scope.info.app.components) return


        for (var i = 0; i < $scope.info.app.components.length; i++) {
            var url = window.decodeURIComponent($location.search().select)
            if ($scope.info.app.components[i].url == url) {
                // $scope.selected = $scope.babies[i]
                $anchorScroll();
                $scope.shaked = $scope.info.app.components[i]
                $location.hash('component_' + url);
                break
            }
        }
    }
}