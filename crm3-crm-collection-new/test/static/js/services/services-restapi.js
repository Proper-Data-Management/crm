/**
 * Created by eldars on 15.03.2016.
 */

MetronicApp.service('RestApiService', function($http,$rootScope) {

    var _chatSocket = false;

    var chatSocket = function(url){
        if (!_chatSocket) {
            //console.log("create chat socket" + url);
            _chatSocket = new WebSocket(url);
        }
        return _chatSocket;
    }
    if ($rootScope.isMobile) {
        var baseHost = $rootScope.uri+"/restapi/"
    }else{
        var baseHost = "/restapi/"
    }
    var post = function (uri,data){
        return  $http.post(baseHost+uri,data);
    }
    var get = function (uri){
        return  $http.get(baseHost+uri);
    }
    return {
        post:post,
        get:get,
        chatSocket:chatSocket

    };

});

