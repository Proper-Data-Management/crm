'use strict';
MetronicApp.controller('SimpleTableController', function($rootScope, $scope, $http, $timeout,UIService, DMLService,$filter,$location,RestApiService, FileUploader,PubSub) {

    UIService.bindUITools($scope);
    $scope.selectedAddingField = {id:0,name:"SelectPlease"};

    $scope.getTotal = function(column){

        //return 555;
        var total = 0;

        $scope.rowCollection.forEach(function(item, i, arr) {
            var val = item[column] * 1;
            total += val;
        });
        //
        //for(var i = 0; i < $scope.rowCollection.length; i++){
        //    var column = $scope.rowCollection[i][column]*1;
        //    total += column;
        //}
        return total;

    }

    $scope.simpleRestPost=function (url,data){
        Metronic.startPageLoading();
        RestApiService.post(url,data).
        success(function(data) {
            Metronic.stopPageLoading();
            if (data.errorText) {
                alert(data.errorText);
            }else{
                alert("OK");
                $scope.bind();
            }
            console.log(data);
        });
    }

    $scope.getQueryParams = function (qs) {
        qs = qs.split('+').join(' ');

        var params = {},
            tokens,
            re = /[?&]?([^=]+)=([^&]*)/g;

        while (tokens = re.exec(qs)) {
            params[decodeURIComponent(tokens[1])] = decodeURIComponent(tokens[2]);
        }

        return params;
    }

    $scope.searchURL = "";
    $scope.selectRecordsState = false;
    $scope.rowCollection = [];
    $scope.pageCount = 0;
    $scope.currentPage = 1;
    $scope.perPage = 25;
    $scope.table_name = "accounts";
    $scope.filter = {};//($location.search()).filter;
    $scope.filter.filterDate1 = new Date();
    $scope.filter.filterDate2 = new Date();

    $scope.init = function(opt){
        $scope.pageCode = opt.pageCode;
        $scope.entityId = opt.entityId;
        $scope.exportXMLFields = opt.exportXMLFields;
        //alert($scope.pageCode);
        $scope.currentPage = opt.currentPage;
        $scope.perPage = opt.perPage;
        $scope.table_name = opt.table_name;


        $scope.selectRecordsState =  (opt.selectRecordsState) ? opt.selectRecordsState : false;
        if (opt.filter) {
            $scope.filter = opt.filter;
        }


        if (opt.filter_set) {
            $scope.filterSetCode =opt.filter_set;


            var itemCode ="filter/"+$rootScope.sessioninfo.id+"/"+$scope.filterSetCode;
            var itemSortField = "sortField/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;
            var itemSortFieldAsc = "sortFieldAsc/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;

            var filterSet = localStorage.getItem(itemCode);

            $scope.sortField = localStorage.getItem(itemSortField);
            $scope.sortFieldAsc = localStorage.getItem(itemSortFieldAsc) === 'true';

            RestApiService.get('query/get?code=filter_sets&perpage=3&page=1').
            success(function(data) {
                $scope.filterSets = data.items;

                $scope.filterSets.forEach( function(item, i, arr){
                    RestApiService.get("query/get?code=get_filter_dtls_by_code&param1=" + item.code).
                    success(function (data) {
                        //console.log(data, "filter");
                        $scope.filterSets[i].filterSet = data.items;
                        $scope.bindWithFilter(false,false);
                    });

                });


            });

            if (!filterSet){
                RestApiService.get("query/get?code=get_filter_dtls_by_code&param1=" + $scope.filterSetCode).
                success(function (data) {
                    //console.log(data, "filter");
                    $scope.filterSet = data.items;
                    $scope.bindWithFilter(false,false);
                });
            }else{
                //$scope.filterSet = JSON.parse(filterSet);
                $scope.setFilterSetFromString(filterSet);
                $scope.bindWithFilter(false,true);
            }

        }else{
            $scope.bind();
        }



    }

    $scope.selectRecords = function (){
        $scope.selectRecordsState = true;
    }

    $scope.deselectRecords = function (){
        $scope.selectRecordsState = false;
    }

    $scope.getSelectRecordState = function (){
        return $scope.selectRecordsState;
    }

    $scope.bindPage = function (inPage,inPerpage){
        $scope.currentPage = inPage;
        $scope.perPage = inPerpage;
        $scope.bind();
    }

    $scope.resetFilter = function (){
        if ($scope.filterSetCode && $rootScope.sessioninfo.id) {
            //alert("test");
            var itemCode = "filter/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;
            var itemSortField = "sortField/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;
            var itemSortFieldAsc = "sortFieldAsc/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;

            console.log(itemCode);
            $scope.sortField == null;
            localStorage.setItem(itemCode,"");
            localStorage.setItem(itemSortField,null);
            localStorage.setItem(itemSortFieldAsc,null);
            location.reload();
        }
    }

    $scope.changeFunc = function(fs){

        //fs.visible = false;
        console.log("changefunc",fs);

        $scope.filterSets.forEach(function(fss, ifs, arrs) {

            fss.filterSet.forEach(function(item, i, arr) {

            if  (item.code == fs.idLink.code) {
                item.input_code = fs.input_code;
                console.log(item.input_code,"$scope.filterSet[i].input_code")
            }
            if  (item.code == fs.idLink.code && fs.is_need_data == 0) {
                if (!item.value) {
                    item.value = {}
                }
                //$scope.filterSet[i].value.text = "-"
                item.hidden = true;
            }else{
                if (!item.value) {
                    item.value = {}
                }
                item.hidden = false
            }
                //$scope.filterSets[ifs].filterSet[i] =item;
            //alert( i + ": " + item + " (массив:" + arr + ")" );
        });
    });

    }

    $scope.delField = function(fs){

        $scope.filterSet.forEach(function(item, i, arr) {
            if  (item.code == fs.code) {
                $scope.filterSet[i].visible = false;
                $scope.filterSet[i].funcId = null;
                $scope.filterSet[i].value = null;
            }
            //alert( i + ": " + item + " (массив:" + arr + ")" );
        });

    }

    $scope.addField = function(fs){

        console.log(fs);
        var positiveArr = $scope.filterSet.filter(function(item) {
            return item.code == fs.code;
        });

        $scope.filterSet.forEach(function(item, i, arr) {
            if  (item.code == fs.code) {
                $scope.filterSet[i].visible = true;
            }
            //alert( i + ": " + item + " (массив:" + arr + ")" );
        });

    }
    $scope.searchByTitle = function(text){
        //alert($scope.filterText);
        $scope.searchURL="flt$title$like$=%25"+text+"%25";
        $scope.bind();
    }

    $scope.setFilter = function(filter){
        $scope.filter = filter;
        $scope.bind();
    }

    $scope.universalSearch = function(filter){
        //alert($scope.filterText);

        $scope.bind();
    }

    $scope.searchByName = function(text){
        //alert($scope.filterText);
        $scope.searchURL="flt$name="+text;
        $scope.bind();
    }

    $scope.saveFilter = function(filterSet){
        //alert("SAVING FILTER");
        Metronic.startPageLoading();
        var itemCode ="filter/"+$rootScope.sessioninfo.id+"/"+$scope.filterSetCode;
        var itemSortField = "sortField/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;
        var itemSortFieldAsc = "sortFieldAsc/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetCode;

        //console.log("itemCode",itemCode);
        //localStorage.setItem(itemCode,JSON.stringify(filterSet));
        localStorage.setItem(itemCode,$scope.getFilterSetAsString());
        localStorage.setItem(itemSortField,$scope.sortField);
        localStorage.setItem(itemSortFieldAsc, $scope.sortFieldAsc);
    }

    $scope.buildFilter = function(load){
        //alert($scope.filterSets);
        if ($scope.filterSets) {

            $scope.searchURL = "";
            var it = 0;

            $scope.filterSets.forEach(function(fss, fss_i, arr1) {
                if (fss.filterSet) {
                    fss.filterSet.forEach(function (item, i, arr2) {


                        item.visible = false;

                        if (item.data_type == "reference" || item.data_type == "double" || item.data_type == "integer"
                            || item.data_type == "varchar"
                            || item.data_type == "timestamp"
                            || item.data_type == "datetime"
                            || item.data_type == "date"
                            || item.data_type == "serial"
                        ) {
                            item.hidden = false;
                            if (item.is_advanced != 1) {
                                item.visible = true;
                            } else {
                                item.visible = false;
                            }
                            it++;
                        }

                        if (!item.funcId && item.data_type == "varchar") {
                            //item.funcId = 2;
                        }

                        if (it % 2 == 0) {
                            item.class = "";
                            //item.style= "background:yellow";
                        } else {
                            item.class = "";
                            //item.style= "background:blue";
                        }

                        if (item.data_type == "reference") {
                            RestApiService.get("list/simple/get?code=" + item.entity_link).success(
                                function (data) {
                                    item.lookup = data;
                                }
                            );
                        }


                        if (item.data_type == "serial") {
                            RestApiService.get("list/simple/get?code=" + item.entity_code).success(
                                function (data) {
                                    item.lookup = data;
                                }
                            );
                        }
                        var prefix = "flt$";

                        if (item.data_type == 'varchar2' || item.data_type == 'longtext2' || item.data_type == 'text2' && (item.value && item.value.text != "" || item.hidden  )) {
                            if (!item.value && typeof item.funcId != "undefined" && item.is_need_data != 1) {
                                $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=&"
                            } else if (item.value && typeof item.value.text != 'undefined' && typeof item.funcId != "undefined") {
                                item.visible = true;
                                $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.text + "&";
                            }

                        }
                        else if (
                            (item.data_type == 'varchar' ||
                                item.data_type == 'timestamp' ||
                                item.data_type == 'date' || item.data_type == 'integer' ||
                                item.data_type == 'double'
                            ) && (item.value)) {


                            if ((item.input_code == "date_interval") && (item.value) && !load) {
                                item.value.from_date = $filter('date')(item.value.from_date_mdl, "yyyy-MM-dd");
                                item.value.to_date = $filter('date')(item.value.to_date_mdl, "yyyy-MM-dd 23:59:59");
                            } else {
                                if (item.value.from_date != null && item.value.to_date != null) {
                                    item.value.from_date_mdl = new Date(item.value.from_date);
                                    item.value.to_date_mdl = new Date(item.value.to_date);
                                }
                            }

                            if (
                                typeof item.funcId != "undefined" &&

                                (
                                item.value.from_value != null && item.value.to_value != null
                                ||
                                item.value.text != null
                                ||
                                item.value.from_date != null && item.value.to_date != null || item.is_need_data != 0 )) {

                                item.visible = true;

                                console.log("item.funcId", item.funcId)
                                console.log("item.value.text", item.value.text)
                                console.log("typeof item.funcId", typeof item.funcId)
                                //$scope.searchURL += prefix + item.code + "$gteq$=" + item.value.from_date + "&";
                                //item.funcId
                                console.log("item.is_need_data", item.is_need_data)
                                if (item.input_code == "date_interval" && item.is_need_data != 0 && typeof item.value.from_date != 'undefined' && typeof item.value.to_date != 'undefined') {
                                    $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.from_date + "|" + item.value.to_date + "&";
                                } else if (item.input_code == "number_interval" && item.is_need_data != 0 && typeof item.value.from_value != 'undefined' && typeof item.value.to_value != 'undefined') {
                                    $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.from_value + "|" + item.value.to_value + "&";
                                } else if (item.input_code == "input_text" && item.is_need_data != 0 && typeof item.value.text != 'undefined' && typeof item.value.text != '') {
                                    $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.text + "&";
                                } else {
                                    $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=&";
                                }
                            }
                        }
                        else if ((item.data_type == 'reference') && (item.value)) {
                            //alert(item.value.selected_references_mdl);
                            if (item.value.selected_references_mdl != null) {
                                var values = "0";
                                var cnt = 0;
                                item.value.selected_references_mdl.forEach(function (itm, i, arr) {
                                    //alert( i + ": " + item + " (массив:" + arr + ")" );
                                    values += "," + itm.id;
                                    cnt++;
                                });
                                if (cnt > 0) {
                                    item.visible = true;
                                    $scope.searchURL += prefix + item.code + "$in$=(" + values + ")&";
                                }
                            }
                        } else if ((item.data_type == 'serial') && (item.value)) {
                            //alert(item.value.selected_references_mdl);
                            if (item.value.selected_references_mdl != null) {
                                var values = "0";
                                var cnt = 0;
                                item.value.selected_references_mdl.forEach(function (itm, i, arr) {
                                    //alert( i + ": " + item + " (массив:" + arr + ")" );
                                    values += "," + itm.id;
                                    cnt++;
                                });
                                if (cnt > 0) {
                                    item.visible = true;
                                    $scope.searchURL += prefix + item.code + "$in$=(" + values + ")&";
                                }
                            }
                        }

                        $scope.filterSets[fss_i].filterSet[i] = item;
                        console.log("www");
                    });
                }
            });
            if (!load) {
                $scope.saveFilter($scope.filterSets);
            }
        }
    }

    $scope.buildFilterOld = function(load){
        if ($scope.filterSet) {
            $scope.searchURL = "";
            var it = 0;
            $scope.filterSet.forEach(function(item, i, arr) {

                item.visible = false;

                if (item.data_type == "reference" || item.data_type == "double" || item.data_type == "integer"
                    || item.data_type == "varchar"
                    || item.data_type == "timestamp"
                    || item.data_type == "datetime"
                    || item.data_type == "date"
                    || item.data_type == "serial"
                ) {
                    item.hidden = false;
                    if (item.is_advanced != 1){
                        item.visible = true;
                    }else{
                        item.visible = false;
                    }
                    it++;
                }

                if (!item.funcId && item.data_type == "varchar") {
                    //item.funcId = 2;
                }

                if (it % 2 == 0) {
                    item.class = "";
                    //item.style= "background:yellow";
                }else{
                    item.class = "";
                    //item.style= "background:blue";
                }

                if (item.data_type == "reference") {
                    RestApiService.get("list/simple/get?code=" + $scope.filterSet[i].entity_link).success(
                        function (data) {
                            $scope.filterSet[i].lookup = data;
                        }
                    );
                }


                if (item.data_type == "serial") {
                    RestApiService.get("list/simple/get?code=" + $scope.filterSet[i].entity_code).success(
                        function (data) {
                            $scope.filterSet[i].lookup = data;
                        }
                    );
                }
                var prefix = "flt$";

                if (item.data_type=='varchar2'||item.data_type=='longtext2' || item.data_type=='text2'  && (item.value && item.value.text!="" || item.hidden  ) ) {
                    if (!item.value && typeof item.funcId != "undefined" && item.is_need_data != 1){
                        $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=&"
                    }else if (item.value  && typeof item.value.text != 'undefined' && typeof item.funcId != "undefined") {
                        item.visible = true;
                        $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.text + "&";
                    }

                }
                else if (
                    (item.data_type=='varchar' ||
                    item.data_type=='timestamp' ||
                    item.data_type=='date'|| item.data_type=='integer'||
                    item.data_type=='double'
                    ) && (item.value)) {


                    if ((item.input_code=="date_interval") && (item.value) &&!load) {
                        item.value.from_date = $filter('date')(item.value.from_date_mdl, "yyyy-MM-dd");
                        item.value.to_date = $filter('date')(item.value.to_date_mdl, "yyyy-MM-dd 23:59:59");
                    }else{
                        if (item.value.from_date!=null && item.value.to_date!=null ){
                            item.value.from_date_mdl = new Date(item.value.from_date);
                            item.value.to_date_mdl = new  Date(item.value.to_date);
                        }
                    }

                    if (
                        typeof item.funcId != "undefined" &&

                        (
                        item.value.from_value!=null && item.value.to_value!=null
                            ||
                        item.value.text!=null
                            ||
                        item.value.from_date!=null && item.value.to_date!=null || item.is_need_data != 0 )) {

                        item.visible = true;

                        console.log("item.funcId",item.funcId)
                        console.log("item.value.text",item.value.text)
                        console.log("typeof item.funcId",typeof item.funcId)
                        //$scope.searchURL += prefix + item.code + "$gteq$=" + item.value.from_date + "&";
                        //item.funcId
                        console.log("item.is_need_data",item.is_need_data)
                        if  (item.input_code=="date_interval" && item.is_need_data != 0 && typeof item.value.from_date!='undefined' && typeof item.value.to_date!='undefined') {
                            $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.from_date + "|" + item.value.to_date + "&";
                        } else if  (item.input_code=="number_interval"  && item.is_need_data != 0 && typeof item.value.from_value!='undefined' && typeof item.value.to_value!='undefined') {
                            $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.from_value + "|" + item.value.to_value + "&";
                        }else if  (item.input_code=="input_text"  && item.is_need_data != 0 && typeof item.value.text!='undefined' && typeof item.value.text!='') {
                            $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.text + "&";
                        }else{
                            $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=&";
                        }
                    }
                }
                else if ((item.data_type=='reference') && (item.value)) {
                    //alert(item.value.selected_references_mdl);
                    if (item.value.selected_references_mdl!=null) {
                        var values = "0";
                        var cnt = 0;
                        item.value.selected_references_mdl.forEach(function(itm, i, arr) {
                            //alert( i + ": " + item + " (массив:" + arr + ")" );
                            values += ","+itm.id;
                            cnt++;
                        });
                        if (cnt>0) {
                            item.visible = true;
                            $scope.searchURL += prefix + item.code + "$in$=(" + values + ")&";
                        }
                    }
                } else if ((item.data_type=='serial') && (item.value)) {
                    //alert(item.value.selected_references_mdl);
                    if (item.value.selected_references_mdl!=null) {
                        var values = "0";
                        var cnt = 0;
                        item.value.selected_references_mdl.forEach(function(itm, i, arr) {
                            //alert( i + ": " + item + " (массив:" + arr + ")" );
                            values += ","+itm.id;
                            cnt++;
                        });
                        if (cnt>0) {
                            item.visible = true;
                            $scope.searchURL += prefix + item.code + "$in$=(" + values + ")&";
                        }
                    }
                }
            });
            if (!load) {
                $scope.saveFilter($scope.filterSet);
            }
        }
    }

    $scope.bindWithFilter = function(save,load){
        if (save) {
            $scope.buildFilter(load);
        }


        $scope.bind();
    }



    $scope.prepareExportXML = function (){
        var dt = "";
        $scope.exportXMLFieldsArr = $scope.exportXMLFields.split(",");
        if ($scope.exportCollection && $scope.exportXMLFieldsArr) {
            $scope.exportCollection.forEach(function (item, i, arr) {
                var s = {};
                //console.log("$scope.exportXMLFields",$scope.exportXMLFields);
                $scope.exportXMLFieldsArr.forEach(function (xmlKey, xmlI, xmlArr) {
                    s[xmlKey] = item[xmlKey];
                });

                dt += "<item>" + json2xml(s) + "</item>";
            });
        }

        $scope.exportXMLData = 'data:application/xml;charset=UTF-8,'+"<?xml version=\"1.0\" encoding=\"UTF-8\"?><items>"+dt+"</items>";
    }

    $scope.exportXML = function(){

        Metronic.startPageLoading();
        $scope.buildFilter(true);
        RestApiService.get('query/get?code='+$scope.table_name+'&'+$scope.searchURL).
        success(function(data) {
            Metronic.stopPageLoading();
            $scope.exportCollection = data.items;
            console.log("$scope.exportCollection",$scope.exportCollection);
            if (data.error==2){
                location.href="/auth/logout";
            }
            $scope.prepareExportXML();
        });
    }

    $scope.deleteUserFilter = function (flt){
        if (confirm("Are you sure?")){
            //console.log("delete");

            $http.post('../restapi/update_v_1_1', {items: [ {table_name:"filter_save_users", action:"delete",values:[{id:flt.id}]}    ]}).
            success(function (data) {
                if (data.error_text!="OK"){
                    alert(data.error_text);
                }
                $scope.bind();
            });

        }
    }
    $scope.loadUserFilter = function (flt){
        $scope.userFilter = flt.url;
        $scope.setFilterSetFromString(flt.filterset);
        $scope.bind();
    }
    $scope.bind = function (){

        Metronic.startPageLoading();
        $scope.buildFilter(true);





        RestApiService.get('query/get?code=entity_acts&param1='+$scope.entityId).
        success(function(data) {
            $scope.entityActions = data.items;
        });


        RestApiService.get('query/get?code=entity_tools&param1='+$scope.entityId).
        success(function(data) {
            $scope.entityTools = data.items;
        });

        RestApiService.get('query/get?code=entity_global_acts').
        success(function(data) {
            $scope.entityGlobalActions = data.items;
        });

        RestApiService.get('query/get?code=filter_save_users_my&param1='+$scope.entityId).
        success(function(data) {
            $scope.filterSaveUsersMy = data.items;
        });

        $scope.searchURLWithSorting = $scope.searchURL;
        if ($scope.sortField) {
            $scope.searchURLWithSorting = $scope.searchURLWithSorting + "orderBy=" + $scope.sortField+"&"
        }
        if ($scope.sortFieldAsc === true) {
            $scope.searchURLWithSorting = $scope.searchURLWithSorting + "orderAsc=1"
        } else if ($scope.sortFieldAsc === false) {
            $scope.searchURLWithSorting = $scope.searchURLWithSorting + "orderAsc=0"
        }

        RestApiService.get('query/get?code='+$scope.table_name+'&page='+$scope.currentPage+'&perpage='+$scope.perPage+"&"+$scope.searchURLWithSorting).
        success(function(data) {
            Metronic.stopPageLoading();
            $scope.rowCollection = data.items;
            $scope.allCount = data.allCount;
            $scope.pageCount = data.pageCount;
            if (data.error==2){
                location.href="/auth/logout";
            }
        });
    }

    var sub = PubSub.subscribe('SimpleTableController.bind', $scope.bind);

    $scope.pubSubPublish = function (str,params){
        PubSub.publish(str, params);
    }


    $scope.getSelectedIDsUrl = function(){
        return $scope.getSelectedIDs().join(",");

    }
    $scope.getSelectedIDs = function (){

        var idValues = [];
        if ($scope.rowCollection){

            $scope.rowCollection.forEach(function (item, i, arr) {
                if (item.selected) {
                    idValues.push(item.id * 1);
                }
            });
            return idValues;
        }else{
            return [];
        }
    }

    $scope.deleteSelectedRecord = function (){
        //alert("test");

        if (!confirm($filter('translate')('Delete Selected Records?'))){
            return;
        }
        var deleteValues = [];
        $scope.rowCollection.forEach(function(item, i, arr) {
            if (item.selected) {
                deleteValues.push({id: item.id});
                //console.log("Deleted " + item.id);
            }
        });
        $http.post('../restapi/update_v_1_1', {items: [ {table_name:$scope.table_name, action:"delete",values:deleteValues}    ]}).
        success(function (data) {
            if (data.error_text){
                alert(data.error_text);
            }
            $scope.bind();
        });
    }

    $scope.setFilterSetFromString = function(str){
        $scope.filterSet = JSON.parse(str);
    }

    $scope.getFilterSetAsString = function(){
        var fs = $scope.filterSet;
        fs.forEach(function(item, i, arr) {
          fs[i].lookup = [];
        });
        return JSON.stringify(fs);
    }
    $scope.exportAll=function(){
        alert("export +"+$scope.table_name);
    }

    $scope.sortBy = function(field){
        $scope.sortField = field;
        if ($scope.sortFieldAsc == null) {
            $scope.sortFieldAsc = false;
        }
        $scope.sortFieldAsc = !$scope.sortFieldAsc;

        $scope.bind();
        $scope.saveFilter($scope.filterSet);
        //console.log("test");
    }

    $scope.removeAll=function(tableName){
        if (confirm($filter('translate')('Delete All Records?'))){
            DMLService.removeAll(tableName).success(function (data) {
                alert(data.ok_text);
            })
        }
    }

    $scope.refreshTest2 = function($select){
        var search = $select.search,
            list = angular.copy($select.items),
            FLAG = -1;
        //remove last user input
        list = list.filter(function(item) {
            return item.id !== FLAG;
        });

        if (!search) {
            //use the predefined list
            $select.items = list;
        }
        else {
            //manually add user input and set selection
            var userInputItem = {
                id: FLAG,
                description: search
            };
            $select.items = [userInputItem].concat(list);
            $select.selected = userInputItem;
        }
    }

    $scope.refreshSelect = function($select,table){
        console.log("refreshSelect",$select);
        if ($select.search.length>=4) {
            RestApiService.get("list/simple/get?code=" + table+"&contains="+$select.search).success(
                function (data) {
                    $select.items = data;
                }
            );
        }

        //$select.items.push({id:-1,title:"vata emes"});
        //return $select;
    }

});