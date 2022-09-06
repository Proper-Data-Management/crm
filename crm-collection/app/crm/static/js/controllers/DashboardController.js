'use strict';

MetronicApp.controller('DashboardController', function($rootScope, $scope, $http, $timeout) {
    $scope.$on('$viewContentLoaded', function() {   
        // initialize core components
        Metronic.initAjax();



    $scope.runWidget = function runWidget(config){
        $http.get(config.url+config.name).
            success(function (data) {
                $scope[config.name] = data
            });
    }

    $scope.runWidget({name:'widget_db_sales_summary',url:'../restapi/widget/'});

        $scope.bind();

    });

    // set sidebar closed and body solid layout mode
    $rootScope.settings.layout.pageBodySolid = true;
    $rootScope.settings.layout.pageSidebarClosed = false;




    $scope.bind = function (){
        $http.get("../restapi/detail?code=dashboard&id="+$rootScope.sessioninfo.default_dashboard_id).
        success(function (data) {
            //alert(data);
            $scope.dashboard_rows=data.dashboard_rows;
            $scope.dashboard_cols=data.dashboard_cols;
        });
    }



});