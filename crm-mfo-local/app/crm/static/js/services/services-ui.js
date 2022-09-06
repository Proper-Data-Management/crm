/**
 * Created by eldars on 04.01.2016.
 */

MetronicApp.service('UIService', function($http,$rootScope,PubSub) {

   var getCurrentPage = function(){
       alert("test");
   }

   var editPage = function(pageCode){
        $http.get('../restapi/query/get?code=pageByCode&param1='+pageCode).
        success(function(data) {
            location.href="#/settings/pagedetails/"+data.items[0].id;
        });
    };

    var editQuery = function(query){
        $http.get('../restapi/query/get?code=queryByCode&param1='+query).
        success(function(data) {
            location.href="#/settings/querydetails/"+data.items[0].id;
        });
    }

    var generateGUID = function () {
        function s4() {
            return Math.floor((1 + Math.random()) * 0x10000)
                .toString(16)
                .substring(1);
        }
        return s4() + s4() + '-' + s4() + '-' + s4() + '-' +
            s4() + '-' + s4() + s4() + s4();
    }

    var bindUITools = function($theScope){
        if ($rootScope.session_roles && $rootScope.session_roles.admin) {
            $theScope.editQuery = editQuery;
            $theScope.generateGUID = generateGUID;
            $theScope.getCurrentPage = getCurrentPage;
        }
    }

    return {
        bindUITools:bindUITools
    };

});

