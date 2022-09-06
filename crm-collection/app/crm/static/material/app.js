var MaterialApp = angular.module("MaterialApp", ['ngMaterial', 'ngMessages']);


MaterialApp.config(function($mdThemingProvider) {
    $mdThemingProvider.theme('docs-dark', 'default')
        .primaryPalette('yellow')
        .dark();
});

MaterialApp.controller('AppController', ['$scope', '$rootScope', function($scope, $rootScope) {

}]);

