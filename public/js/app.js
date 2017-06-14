angular.module("flow", [
    'ngRoute',
    'ui.router',
    'ngSanitize',
    'ui.bootstrap',
    'relativeDate'
])
angular.module("flow").constant('constants', {
    version: "1.0.2"
})
// configure our routes
angular.module("flow").config(function ($routeProvider, $stateProvider, $urlRouterProvider) {

    $urlRouterProvider.otherwise("/")

    $stateProvider
        .state('index', {
            url: "/",
            views: {
                "single": {
                    templateUrl: 'views/components.html',
                    controller: placesController,
                    resolve: placesController.resolve
                }
            }
        })
        .state('component', {
            url: "/listen/:component",
            views: {
                "single": {
                    templateUrl: 'views/component.html',
                    controller: placeController,
                    resolve: placesController.resolve
                }
            }
        })
})
angular.module("flow").run(function ($rootScope) {

})

angular.module("flow").controller('mainController', function ($scope) {

})