MetronicApp.service('UISimpleTable', function($http) {

    var selectRecordsState = false;
    var rowCollection = [];
    var pageCount = 0;


    var currentPage = 1;
    var perPage = 25;
    var table_name = "accounts";

    var init = function(opt){
        currentPage = opt.currentPage;
        perPage = opt.perPage;
        table_name = opt.table_name;
        selectRecordsState =  (opt.selectRecordsState) ? opt.selectRecordsState : false;
    }
    var selectRecords = function (){
        selectRecordsState = true;
    }

    var deselectRecords = function (){
        selectRecordsState = false;
    }

    var getSelectRecordState = function (){
        return selectRecordsState;
    }

    var bindPage = function (inPage,inPerpage){
        currentPage = inPage;
        perPage = inPerpage;
        bind();
    }

    var bind = function (){
        $http.get('../restapi/query/get?code='+table_name+'&page='+currentPage+'&perpage='+perPage).
        success(function(data) {
            /*if (data.session_info.user_id==0) {
             location.href = "/auth/logout";
             return;
             }*/
            rowCollection = data.items;
            pageCount = data.pageCount;
            //alert(pageCount);
            console.log(data.items);
        });
    }

    var getRowCollection = function (){
        return rowCollection;

    }
    var deleteSelectedRecord = function (){
        //alert("test");

        var deleteValues = [];
        rowCollection.forEach(function(item, i, arr) {
            if (item.selected) {
                deleteValues.push({id: item.id});
                console.log("Deleted " + item.id);
            }
        });
        $http.post('../restapi/update_v_1_1', {items: [ {table_name:table_name, action:"delete",values:deleteValues}    ]}).
        success(function (data) {
            bind();
        });
    }

    function getCurrentPage(){
        return currentPage;
    }
    function getPerPage(){
        return perPage;
    }
    function getPageCount(){
        return pageCount;
    }
    return {
        init: init,
        selectRecords: selectRecords,
        deselectRecords: deselectRecords,
        getSelectRecordState:getSelectRecordState,
        deleteSelectedRecord: deleteSelectedRecord,
        bind:bind,
        bindPage:bindPage,
        getRowCollection:getRowCollection,
        currentPage:getCurrentPage,
        perPage:getPerPage,
        pageCount:getPageCount
    };

});