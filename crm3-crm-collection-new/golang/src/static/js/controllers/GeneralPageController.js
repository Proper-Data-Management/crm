
/* Setup general page controller */
MetronicApp.controller('GeneralPageController', ['$rootScope', '$scope', 'settings','PubSub',


    function($rootScope, $scope, settings,PubSub) {

        $scope.getCurrentPage = function(){
            return '555';
        }

        var sub = PubSub.subscribe('GeneralPageController.getCurrentPage', $scope.getCurrentPage);


        $scope.$on('$viewContentLoaded', function() {
            // initialize core components

            //console.log(settings,"settings");
            Metronic.initAjax();
            // set default layout mode
            $rootScope.settings.layout.pageBodySolid = false;
            $rootScope.settings.layout.pageSidebarClosed = false;
            $rootScope.date = new Date();

        });
}]);

