/**
 * Created by eldars on 04.01.2016.
 */

MetronicApp.service('DMLService', function($http) {

    var update = function (items){
        return  $http.post('../restapi/update_v_1_1', {items:  items});
    }

    var removeAll = function(tableName){
        return $http.get('../restapi/removeall?code='+tableName);
    }
    return {
        update:update,
        removeAll:removeAll
    };

});

