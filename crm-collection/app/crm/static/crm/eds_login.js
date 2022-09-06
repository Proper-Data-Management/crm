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
                $translateProvider.translations('en', data["en"]).translations('ru', data["ru"]).translations('kk', data["kk"]);
                $translateProvider.preferredLanguage('ru');

            });
        }
    });


LoginApp.run(['$rootScope', '$translate', '$log', function ($rootScope, $translate, $log) {

    $rootScope.changeLanguage = function (langKey) {
        $translate.use(langKey);
    };
}]);






LoginApp.controller('LoginController', function($scope, $http, $timeout,NCALayer) {


	$scope.textToSign = "555";
 
 
 	$http
		.get("../../restapi/services/run/login_logo")
		.success(function(data) {
			 $scope.logo_url = data.url;
		});	
   

    NCALayer.showPassword = function (show) {
        if (show) {
			$scope.showPassword = true;
            //$scope.step = $scope.steps.inputPassword;
        } else {
			$scope.showPassword = false;
            //$scope.step = $scope.steps.selectCert;
        }
        
        if(!$scope.$$phase) {
            $scope.$apply();
        }
    };
	
	
    NCALayer.postGetSubjectDN = function  (answer) {
		//alert('postGetSubjectDN');
        //$scope.hideError();
        var arr = answer.result.split(',');
        for(var i = 0; i < arr.length; i++) {
            var keyVal = arr[i].split('=');

            if(keyVal[0] == 'SERIALNUMBER') {
                $scope.person[keyVal[0]] = keyVal[1].replace(/^IIN/,'');

            } else if(keyVal[0] == 'OU') {
                $scope.person[keyVal[0]] = keyVal[1].replace(/^BIN/,'');
            } else {
                $scope.person[keyVal[0]] = keyVal[1];
            }
            
            if(!$scope.$$phase) {
                $scope.$apply();
            }
            
			
            console.log(keyVal[0] +  ' = ' + $scope.person[keyVal[0]] + ", " + keyVal[1]);
        }   
		
    };
	
    NCALayer.postGetDate = function  (answer,dateName) {
		
		
        $scope.person[dateName] = answer.result.split(' ')[0];
        console.log($scope.person[dateName]);
        $scope.$apply();
		
		
        
        if($scope.person['endDate']) {
            
            if($scope.bin) {
                if($scope.bin == $scope.person['OU']) {
                    NCALayer.showPerson(true);
                } else if($scope.bin == $scope.person['SERIALNUMBER']) {
                    NCALayer.showPerson(true);
                } else {
                    NCALayer.showPerson(true); 
                    NCALayer.showError({ errorCode : 'Поле ИНН/БИН не совпадает. Требуется серийный номер -  ' + $scope.bin, result : -1 });
                }
            } else {
                NCALayer.showPerson(true);
            }
        }
		
		$scope.signXml("<root></root>");
		
    };	
	
	
    NCALayer.postSignXml = function (answer) {
		
		//alert('postSignXml');
        $scope.certificate = answer.result;
        //alert(answer.result);
        
		
        $http
            .post("/restapi/services/run/auth_by_eds",{signxml:$scope.certificate})
            .success(function(data) {
				if (data.output && data.output.redirect_url){
					location.href = data.output.redirect_url;
				}
            });

			
        if(!$scope.$$phase) {
            $scope.$apply();
        }
    };
	
    NCALayer.showPerson = function (show) {
        /*if (show) {
            $scope.step = $scope.steps.showData;
        } else {
            $scope.step = $scope.steps.selectCert;
        }*/
        
        if(!$scope.$$phase) {
            $scope.$apply();
        }
    };
	
	
    $scope.doForget = function(){
        var loginData = {login: $scope.login};
        $http
            .post("/restapi/services/run/send_forget_password_link",loginData)
            .success(function(data) {
                location.href = "login.html?okSendForget=true";
            });
    }
	
	$scope.chooseEDS =  function(){
		//alert('chooseEDS');
		NCALayer.bind($scope);
		$scope.selectSignType();
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

	
	
	$scope.sign = function(password) {
				
			$scope.setNCAPassword(password);
			//$scope.signXml("<root></root>");
		
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