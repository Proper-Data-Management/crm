'use strict';
MetronicApp.controller('SimpleTableController', function($rootScope, $scope, $http, $timeout,UIService, DMLService,$filter,$location,RestApiService, FileUploader,PubSub) {


    UIService.bindUITools($scope);
    $scope.selectedAddingField = {id:0,name:"SelectPlease"};

    $scope.searchMap = [];

    $scope.printExcel = function(incom){
        RestApiService.post("services/run/query_to_xls",
            {
                url:'?code='+$scope.table_name+'&'+$scope.searchURL,
                filename:incom.filename,
                template:incom.template
            }
        ).
        success(function (data) {
            //console.log("data",data);
            location.href=data.url;
        });
    }

    $scope.searchMap2Param =function(){
        var p = "";
        $scope.searchMap.forEach(function(item, i, arr){
            p+=p+"&df_"+item.code+"="+item.value;
        });
        return p;
    }
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
    $scope.selectRecordsState = true;
    $scope.rowCollection = [];
    $scope.pageCount = 0;
    $scope.currentPage = 1;
    $scope.perPage = 25;
    $scope.table_name = "accounts";
    $scope.filter = {};//($location.search()).filter;
    $scope.filter.filterDate1 = new Date();
    $scope.filter.filterDate2 = new Date();

    $scope.init = function(opt){
        $scope.opt = opt;
        //console.log('opt.pageData',opt.pageData );

        $scope.pageCode = opt.pageCode;
        $scope.pageData = opt.pageData;
        $scope.pageId = opt.pageId;
        if (opt.pageData){
            $scope.entityId = opt.pageData.entityId;
            $scope.table_name = opt.pageData.entityCode;
            $scope.query_code = opt.pageData.pageQueryCode;
        }
        $scope.exportXMLFields = opt.exportXMLFields;
        //alert('test');
        //alert($location.search().page);
        if (opt.currentPage){
            $scope.currentPage = opt.currentPage;
        }
        if ($location.search().page)
        {
            $scope.currentPage = $location.search().page?$location.search().page*1:1;
        }

        $scope.perPage = opt.perPage;




        $scope.selectRecordsState =  (opt.selectRecordsState) ? opt.selectRecordsState : true;
        if (opt.filter) {
            $scope.filter = opt.filter;
        }


        if (opt.pageData && opt.pageData.pageFilterSetId) {
            $scope.filterSetId =opt.pageData.pageFilterSetId;

			var filterSet = null;
            var itemCode ="filter/"+$rootScope.sessioninfo.id+"/"+$scope.filterSetId;
            var itemSortField = "sortField/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;
            var itemSortFieldAsc = "sortFieldAsc/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;
			if (!$rootScope.isMobile){	
				var filterSet = localStorage.getItem(itemCode);
				$scope.sortField = localStorage.getItem(itemSortField);
				$scope.sortFieldAsc = localStorage.getItem(itemSortFieldAsc) === 'true';
			}


            RestApiService.get("query/get?code=get_filter_cols_by_id&param1=" + $scope.filterSetId).
            success(function (data) {
                $scope.filterSetCols = data.items;
            });

            if (!filterSet){
                //RestApiService.get("query/get?code=get_filter_dtls_by_id&param1=" + $scope.filterSetId).
				RestApiService.get("services/run/get_filter_dtls_by_id?set_id=" + $scope.filterSetId).
				 
                success(function (data) {
                    //console.log(data, "filter");
                    $scope.filterSet = data.items;
					$scope.filterSet.forEach(
					function (filterItem){
						console.log('filter',filterItem);
						//filterItem.value = {};
						//filterItem.value.selected_references_mdl=[{id:'1',title_short:'Пример'}];
					}

					);
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

    $scope.selectAllRecords = function (){



        $scope.selectRecords();

        $scope.rowCollection.forEach(function(item, i, arr) {
            item.selected = true;
          });

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
        if ($scope.filterSetId && $rootScope.sessioninfo.id) {
            //alert("test");
            var itemCode = "filter/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;
            var itemSortField = "sortField/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;
            var itemSortFieldAsc = "sortFieldAsc/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;

            console.log(itemCode);
            $scope.sortField == null;
			if (!$rootScope.isMobile){
				localStorage.setItem(itemCode,"");
				localStorage.setItem(itemSortField,null);
				localStorage.setItem(itemSortFieldAsc,null);
			}
            location.reload();
        }



    }

    $scope.initFilter = function(filter_set){
        $scope.opt.filter_set = filter_set;
        $scope.init($scope.opt);
    }
    $scope.changeFunc = function(fs){

        //fs.visible = false;
        console.log("changefunc",fs)

        $scope.filterSet.forEach(function(item, i, arr) {

            if  (item.code == fs.idLink.code) {
                $scope.filterSet[i].input_code = fs.input_code;
                console.log($scope.filterSet[i].input_code,"$scope.filterSet[i].input_code")
            }
            if  (item.code == fs.idLink.code && fs.is_need_data == 0) {
                if (!$scope.filterSet[i].value) {
                    $scope.filterSet[i].value = {}
                }
                //$scope.filterSet[i].value.text = "-"
                $scope.filterSet[i].hidden = true
            }else{
                if (!$scope.filterSet[i].value) {
                    $scope.filterSet[i].value = {}
                }
                $scope.filterSet[i].hidden = false
            }
            //alert( i + ": " + item + " (массив:" + arr + ")" );
        });

    }

    $scope.clearField = function(fs){
        $scope.filterSet.forEach(function(item, i, arr) {
            if  (item.code == fs.code) {
                //$scope.filterSet[i].visible = false;
                $scope.filterSet[i].funcId = null;
                $scope.filterSet[i].input_code = null;
                $scope.filterSet[i].value = null;
                $scope.filterSet[i]._selected.name = ""
                //alert($scope.filterSet[i]._selected.name);

            }
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

        //$scope.selectedAddingField = {id:0,name:"SelectPlease"};

    }

    $scope.addField = function(fs){

        //console.log(selectedAddingField);



        var positiveArr = $scope.filterSet.filter(function(item) {
            return item.code == fs.selected.code;
        });

        $scope.filterSet.forEach(function(item, i, arr) {
            if  (item.code == fs.selected.code) {
                $scope.filterSet[i].visible = true;
            }
        });

        //selectedAddingField =null;

        //alert($scope.selectedAddingField.id);

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
        var itemCode ="filter/"+$rootScope.sessioninfo.id+"/"+$scope.filterSetId;
        var itemSortField = "sortField/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;
        var itemSortFieldAsc = "sortFieldAsc/" + $rootScope.sessioninfo.id + "/" + $scope.filterSetId;

        //console.log("itemCode",itemCode);
        //localStorage.setItem(itemCode,JSON.stringify(filterSet));
		if (!$rootScope.isMobile){	
			localStorage.setItem(itemCode,$scope.getFilterSetAsString(filterSet));
			localStorage.setItem(itemSortField,$scope.sortField);
			localStorage.setItem(itemSortFieldAsc, $scope.sortFieldAsc);
		}
    }

    $scope.buildFilter = function(load){
        if ($scope.filterSet) {
            if ($scope.load) {
                $scope.currentPage = 1;
            }
            $scope.searchMap = [];
            $scope.searchURL = "";
            var it = 0;
            $scope.filterSet.forEach(function(item, i, arr) {

                item.visible = false;

                if (item.data_type == "reference" || item.data_type == "double" || item.data_type == "integer"
                    || item.data_type == "varchar"
					|| item.data_type == "longtext"
					|| item.data_type == "text"
                    || item.data_type == "boolean"
                    || item.data_type == "timestamp"
                    || item.data_type == "current_datetime"
                    || item.data_type == "current_and_on_update_datetime"
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

                if (!item.funcId && ( item.data_type == "varchar" || item.data_type == "text" || item.data_type == "longtext" ) ) {
                    //item.funcId = 2;
                }

                if (it % 2 == 0) {
                    item.class = "";
                    //item.style= "background:yellow";
                }else{
                    item.class = "";
                    //item.style= "background:blue";
                }
				
				rolesidlist = ""
				if ($rootScope.sessioninfo && $rootScope.sessioninfo.rolesidlist){
					rolesidlist = $rootScope.sessioninfo.rolesidlist
				}

                if (item.data_type == "reference") {
                    RestApiService.get("list/simple/get?code=" + $scope.filterSet[i].entity_link + "&cacheRolesidlist="+rolesidlist).success(
                        function (data) {
                            $scope.filterSet[i].lookup = data;
                        }
                    );
                }


                if (item.data_type == "serial") {
                    RestApiService.get("list/simple/get?code=" + $scope.filterSet[i].entity_code + "&cacheRolesidlist="+rolesidlist).success(
                        function (data) {
                            $scope.filterSet[i].lookup = data;
                        }
                    );
                }
                var prefix = "flt$";

                if (item.data_type=='varchar2'||item.data_type=='longtext2' || item.data_type=='text2'  && (item.value && item.value.text!="" || item.hidden  ) ) {
                    if (!item.value && typeof item.funcId != "undefined" && item.is_need_data != 1 && item.funcId != null ){
                        $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=&"
                    }else if (item.value  && typeof item.value.text != 'undefined' && typeof item.funcId != "undefined") {
                        item.visible = true;
                        $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.text + "&";
                    }

                }
                else if (
                    (
					item.data_type=='varchar' ||
					item.data_type=='text' ||
					item.data_type=='longtext' ||
                    item.data_type=='boolean' ||
                    item.data_type=='timestamp' ||
                    item.data_type=='current_datetime' ||
                    item.data_type=='current_and_on_update_datetime' ||
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
                        ( typeof item.funcId != "undefined" ||

                        item.data_type == "boolean" && typeof item.value.text!="undefined") &&

                        (
                        item.value.from_value!=null && item.value.to_value!=null
                            ||
                        item.value.text!=null
                            ||
                        item.value.from_date!=null && item.value.to_date!=null || item.is_need_data != 0 )) {

                        item.visible = true;

                        console.log("item.funcId",item.funcId)
                        console.log("item.data_type",item.data_type)
                        console.log("item.value.text",item.value.text)
                        console.log("typeof item.funcId",typeof item.funcId)
                        //$scope.searchURL += prefix + item.code + "$gteq$=" + item.value.from_date + "&";
                        //item.funcId
                        console.log("item.is_need_data",item.is_need_data);

                        if (item.funcId!=null) {
                            if (item.input_code == "date_interval" && item.is_need_data != 0 && typeof item.value.from_date != 'undefined' && typeof item.value.to_date != 'undefined') {
                                $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.from_date + "|" + item.value.to_date + "&";
                            } else if (item.input_code == "number_interval" && item.is_need_data != 0 && typeof item.value.from_value != 'undefined' && typeof item.value.to_value != 'undefined') {
                                $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.from_value + "|" + item.value.to_value + "&";
                            } else if (item.input_code == "input_text" && item.is_need_data != 0 && typeof item.value.text != 'undefined' && typeof item.value.text != '') {
                                $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=" + item.value.text + "&";
                                $scope.searchMap.push({"code": item.attr_code, "value": item.value.text});
                            } else {
                                $scope.searchURL += prefix + item.code + "$func" + item.funcId + "$=&";
                            }
                        }else if (item.data_type == "boolean" && item.value && (item.value.text == "1" || item.value.text == "0") ) {
                                $scope.searchURL += prefix + item.code + "$in$=("+item.value.text+")&";
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
							console.log("item",itm);
                            cnt++;
                        });
                        if (cnt>0) {
                            item.visible = true;
                            $scope.searchURL += prefix + item.code + "$in$=(" + values + ")&";
                        }
                    }
                } else if ((item.data_type == "boolean") && (!item.value) && (item.is_advanced!=1)) {
                    $scope.searchURL += prefix + item.code + "$in$=(0)&";
                }
                //if ($scope.sortField) {
                //    $scope.searchURL = $scope.searchURL + "orderBy=" + $scope.sortField+"&"
                //}
                //if ($scope.sortFieldAsc) {
                //    $scope.searchURL = $scope.searchURL + "orderAsc=1"
                //}


            });
            if (!load) {
                $scope.saveFilter($scope.filterSet);
            }
        }
    }

    $scope.bindWithFilter = function(save,load){
        if (save) {
            $scope.buildFilter(load);
            $scope.currentPage = 1;
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
    $scope.keyPress = function(e) {
        if (e.keyCode == 13){
            $scope.bindWithFilter(true);
        }
    };
    $scope.keyDown = function(e) {

        //if ((e.ctrlKey) && (e.keyCode == 13)){
        if (e.keyCode == 13){
            $scope.bindWithFilter(true);
        }
    };
    $scope.exportXML = function(){

        Metronic.startPageLoading();
        $scope.buildFilter(true);
        RestApiService.get('query/get?code='+$scope.query_code+'&'+$scope.searchURL).
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

    //Быстрый поиск
    $scope.editQF = function(col){
        console.log(col._value2,"value2");
        console.log(col._valueDate1,"_valueDate1");
        console.log(col._valueDate2,"_valueDate2");
        //console.log(col._value2);
        $scope.searchURL = "";
        $scope.filter = null;
        $scope.quickFilter = $scope.filterSetCols.map (function (item)
        {
            if (item.alias == col.alias && item._data_type_code =="reference"){
                if (col._value2){
                    item._value = col._value2.id;
                }


            }

            if (item._value) {

                if (item._data_type_code !="reference") {
                    return "flt$" + item.default_dtl_id + "$like$=%25" + item._value + "%25";
                }else{
                    return "flt$" + item.default_dtl_id + "$eq$=" + item._value;
                }
            }

            if (item._valueDate1 && item._valueDate2) {
                    return "flt$" + item.default_dtl_id + "$func5$="+item._valueDate1+"|"+item._valueDate2+" 23:59:59";
            }

        }).join("&");
        console.log($scope.quickFilter,"$scope.quickFilter");

        $scope.bind();
    }

    //Быстрый поиск
    $scope.QFKeyDown = function(col,e){
        if (e.keyCode != 13) {
            return
        }
        console.log(col._value2,"value2");
        console.log(col._valueDate1,"_valueDate1");
        console.log(col._valueDate2,"_valueDate2");
        //console.log(col._value2);
        $scope.searchURL = "";
        $scope.filter = null;
        $scope.quickFilter = $scope.filterSetCols.map (function (item)
        {
            if (item.alias == col.alias && item._data_type_code =="reference"){
                if (col._value2){
                    item._value = col._value2.id;
                }


            }

            if (item._value) {

                if (item._data_type_code !="reference") {
                    return "flt$" + item.default_dtl_id + "$like$=%25" + item._value + "%25";
                }else{
                    return "flt$" + item.default_dtl_id + "$eq$=" + item._value;
                }
            }

            if (item._valueDate1 && item._valueDate2) {
                    return "flt$" + item.default_dtl_id + "$func5$="+item._valueDate1+"|"+item._valueDate2+" 23:59:59";
            }

        }).join("&");
        console.log($scope.quickFilter,"$scope.quickFilter");
        $scope.currentPage = 1;
        $scope.bind();
    }

    $scope.bind = function (){

        $scope.pageParam = {};
        Metronic.startPageLoading();
        $scope.buildFilter(true);


		if (typeof $scope.pageId != "undefined"){
			RestApiService.get('query/get?code=page_params_bool&param1='+$scope.pageId).
			success(function(data) {
				$scope.pageParamsBool = data.items;
				if ($scope.pageParamsBool) {
					$scope.pageParamsBool.forEach(function (item, i, arr) {
						var storage = "pageParam/" + $rootScope.sessioninfo.id + "/" + item.page_code+"/"+item.code;
						if (!$rootScope.isMobile){
							item.value = localStorage.getItem(storage)==null?item.value:localStorage.getItem(storage);
							$scope.pageParam[item.code]= item.value;
						}
					});
				}
			});
		}
		if (typeof $scope.entityId != "undefined"){

			RestApiService.get('query/get?code=filter_sets_by_entity&param1='+$scope.entityId).
			success(function(data) {
				$scope.filterSets = data.items;
			});
			RestApiService.get('query/get?code=entity_acts_many&param1='+$scope.entityId).
			success(function(data) {
				$scope.entityActions = data.items;
			});
		}


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
        if  ($scope.searchURL){
            $scope.showAdvancedSearch();
        }
        $scope.searchURLWithSorting = $scope.searchURL;
        if ($scope.sortField) {
            $scope.searchURLWithSorting = $scope.searchURLWithSorting + "orderBy=" + $scope.sortField+"&"
        }
        if ($scope.sortFieldAsc === true) {
            $scope.searchURLWithSorting = $scope.searchURLWithSorting + "orderAsc=1"
        } else if ($scope.sortFieldAsc === false) {
            $scope.searchURLWithSorting = $scope.searchURLWithSorting + "orderAsc=0"
        }

        if (!$scope.quickFilter){
            $scope.quickFilter = "";
        }
        RestApiService.get('query/get?code='+$scope.query_code+'&page='+$scope.currentPage+'&perpage='+$scope.perPage+"&"+$scope.searchURLWithSorting+"&"+$scope.quickFilter).
        success(function(data) {

            $scope.needFilter = data.needFilter;
            $scope.error = data.error;

            if ($scope.needFilter == 1) {
                $scope.showAdvancedSearch();
            }

            Metronic.stopPageLoading();
            $scope.rowCollection = data.items;
            $scope.allCount = data.allCount;
            $scope.pageCount = data.pageCount;
            if (data.error==2){
                location.href="/auth/logout";
            }
        });
    }

    PubSub.unsubscribe('SimpleTableController.bind');
    var sub = PubSub.subscribe('SimpleTableController.bind', $scope.bind);

    $scope.pubSubPublish = function (str,params){
        PubSub.publish(str, params);
    }

    $scope.getURLParameter = function(param){
        return Metronic.getURLParameter(param);
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

    $scope.getFilterSetAsString = function(filterSet){
        var fs = filterSet;
        fs.forEach(function(item, i, arr) {
            item.lookup = null;
            if (item._selected){
                item._selected.idLink = null;
            }
        });
        //console.log("fs",filterSet);
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
        //console.log("refreshSelect",$select);
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
    $scope.setShowIsAdvancedCol = function(item){
        var storage = "showIsAdvancedCol/" + $rootScope.sessioninfo.id + "/" +item.id;
		if (!$rootScope.isMobile){	
        localStorage.setItem(storage,item.value);
		}
    }

    $scope.getShowIsAdvancedCol = function(item){
		if (!$rootScope.isMobile){	
			var storage = "showIsAdvancedCol/" + $rootScope.sessioninfo.id + "/" +item.id;
			return localStorage.getItem(storage);
		}else{
			return false;
		}
    }

    $scope.savePageParamsBool = function(item){
		if (!$rootScope.isMobile){	
			$scope.pageParam[item.code]=item.value;
			var storage = "pageParam/" + $rootScope.sessioninfo.id + "/" + item.page_code+"/"+item.code;
			localStorage.setItem(storage,item.value);
		}
    }

    $scope.showAdvancedSearch = function(){

        $scope.advancedSearch = true;
    }

    $scope.hideAdvancedSearch = function(){
        $scope.advancedSearch = false;
    }

    $scope.expTemplateCallBack = function(params){
        if (params.vars && params.vars.file_url){
        location.href = params.vars.file_url;
        }

    }
});