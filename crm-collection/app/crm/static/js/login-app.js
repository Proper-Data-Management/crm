/***
Metronic AngularJS App Main Script
***/

/* Metronic App */
var LoginApp = angular.module("LoginApp", [
    "pascalprecht.translate",


]);




	
$.ajax({
        url: "../../restapi/translates/get",
        dataType: 'json',
        async: false,

        success: function(data) {
            //stuff
            //...
            LoginApp.config(function($translateProvider) {

                if (data && data["en"]){
                    $translateProvider.translations('en', data["en"]).translations('ru', data["ru"]).translations('kk', data["kk"]);
                    $translateProvider.preferredLanguage('ru');
                }

            });
        }
    });


LoginApp.run(['$rootScope', '$translate', '$log', function ($rootScope, $translate, $log) {

    $rootScope.changeLanguage = function (langKey) {
        $translate.use(langKey);
    };
}]);





LoginApp.controller('LoginController', function($scope, $http, $timeout) {

	$http
		.get("../../restapi/services/run/login_logo")
		.success(function(data) {
			 $scope.logo_url = data.url;
		});	
	
    $scope.doForget = function(){
        var loginData = {login: $scope.login};
        $http
            .post("/restapi/services/run/send_forget_password_link",loginData)
            .success(function(data) {
                location.href = "login.html?okSendForget=true";
            });
    }

    $scope.doResetPassword = function(){
        if ($scope.password1 != $scope.password2){
            $scope.passwordMismatch = true;
        }
        var password = {password: $scope.password1,token : Metronic.getParameterByName("token")};
        $http
            .post("/restapi/services/run/reset_password_by_token",password)
            .success(function(data) {
                location.href = "login.html?okResetPassword=true";
            });
    }

 $scope.doLogin = function(){
     var loginData = {login: $scope.login, password: $scope.password, system: "browser"};
     $scope.loginIncorrect = false;
     $http
         .post("../../restapi/login",loginData)
         .success(function(data) {
             if (data.Result === "ok")
             {
                 if (Metronic.getParameterByName("angularjs_redirecturi")!=null
                 && Metronic.getParameterByName("angularjs_redirecturi")!=""
                 ) {
                     data.RedirectURL = Metronic.getParameterByName("angularjs_redirecturi");
                 }
                 location.href = "/static/#"+data.RedirectURL;
             }
             else {
                $scope.loginIncorrect = true;

             }


         });
 }

});