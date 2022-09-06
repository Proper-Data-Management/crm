
//var myMod = angular.module('myMod', ['ngRoute']);


MetronicApp.component('buSelect', {
    template:
    "<div class=\"form-group\">"+
"                <label ng-if=\"!$ctrl.withoutLabel\" translate>{{$ctrl.label}}</label>"+
"                <div class=\"input-group col-md-12 col-xs-12\">"+
"                  <ui-select ng-if=\"$ctrl.withoutSelect!='true' && $ctrl.readonly!='true'\"  style=\"min-width:100px\" ng-change=\"$ctrl.edit($select.selected)\" ng-model=\"$ctrl.model\" theme=\"bootstrap\">"+
"                    <ui-select-match  placeholder=\"{{$ctrl.label}}\">{{$ctrl.rowPrefix}} {{$select.selected.name | translate}} </ui-select-match>"+
"                    <ui-select-choices refresh-delay=\"0\" refresh=\"$ctrl.refreshSelect($select,$ctrl.queryCode)\" repeat=\"item in $ctrl.values | filter: $select.search\">"+
"                      <div style=\"background-color:{{item.background_color}}\" ng-bind-html=\"$ctrl.rowPrefix + ' ' + item.name | translate | highlight: $select.search\"></div>"+
"                    </ui-select-choices>"+
"                  </ui-select>"+
"       <label ng-if=\"$ctrl.withoutSelect=='true'\">{{$ctrl.model.name}}</label>"+
"		<input ng-required ng-if=\"$ctrl.withoutSelect!='true' && $ctrl.readonly=='true' && $ctrl.model.name\" readonly  ng-model=\"$ctrl.model.name\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+
"        		  <span ng-if=\"!$ctrl.withoutClear && !$ctrl.readonly  && $ctrl.idModel\" title=\"{{'clearCustom' | translate }}\" ng-click=\"$ctrl.clear($select.selected)\" class=\"input-group-addon\">"+
"        		    <i class=\"fa fa-remove\"></i>"+
"        		  </span>  "+
"        		  <span ng-if=\"$ctrl.browseUrl && $ctrl.idModel \" title=\"{{'BrowseUrlTitle' | translate }}\" class=\"input-group-addon\">"+
"        		    <a href=\"{{$ctrl.browseUrl}}\"><i class=\"fa fa-pencil\"></i></a>"+
"        		  </span>  "+
"        		  <span ng-if=\"$ctrl.lookupPageCode\" title=\"{{'lookup' | translate }}\" ng-click=\"$ctrl.lookup()\" class=\"input-group-addon\">"+
"        		    <i class=\"fa fa-search\"></i>"+
"        		  </span>  "+
//"        		  <a title=\"{{'Open Entity' | translate }}\"  href=\"#/settings/entitydetails/{{detail._entity_id.id}}\" class=\"input-group-addon\">"+
//"        		    <i class=\"glyphicon glyphicon-chevron-right\"></i>"+
//"        		  </a>                         "+
"                </div>"+
"            </div> "+
        "",
    bindings: {
        data: '@',
        label: '@',
        queryCode: '@',
        value: '=',
        values: '<',
        rowPrefix: '@',
        idModel : '=',
        idLink : '=',
        onChange : '=',
        bindOnInit: '@',
        title: '@',
        withoutLabel: '@',
        browseUrl: '@',
        withoutClear: '@',
        withoutSelect: '@',
        withTranslate: '@',
        cacheData: '@',
        readonly:'@',
        lookupPageCode:'@',
        lookupCallBack:'@',

    },
    controller: function ($scope,RestApiService,PubSub) {



        var self = this;

        self.loaded = false;

        $scope.$watch(
            "$ctrl.queryCode",
            function handleFooChange( newValue, oldValue ) {
                if (typeof self.queryCode !== "undefined" && self.queryCode !== null  && self.loaded /* && typeof self.idModel !== "undefined" && self.idModel !== null */){
                    console.log("self.queryCode1",self.queryCode);;
                    console.log("self.queryCode1",self.idModel);;
                    self.bind();

                }


            }
        );

        $scope.$watch(
            "$ctrl.idModel",
            function handleFooChange( newValue, oldValue ) {
                //console.log(oldValue);
                //console.log(newValue);
                //console.log(oldValue);
                if (typeof self.queryCode !== "undefined" && self.queryCode !== null){
                    //self.idOrigModel = new Date(newValue);
                    //console.log("norm777555 "+self.idModel);
                    //console.log("self.queryCode2",self.queryCode)
                    //console.log("self.queryCode2",self.idModel)
                    self.bind();
                }
            }
        );

        self.$onInit = function() {
            //console.log(self.queryCode);
            if (self.bindOnInit) {
                self.bind();
            }
            if (!self.idModel){
                self.idModel = null;
            }

            //console.log("subscribing to bindBuComponents");
            //PubSub.subscribe("bindBuComponents",self.bind);
        };


        self.edit = function(el){
            console.log("el",el);

            //alert(self.model.id);
            self.idModel = self.model.id;
            self.value = self.model;

            //if (el.id) {
            //    self.idModel = el.id;
            //    self.value = self.idLink;
            //}



            if (typeof self.onChange === "function") {
                el.idLink = self.idLink;
                self.onChange(el);
            }
            //self.onEdit();
        };;

        self.clear = function(el){
            self.model={id: null, name: ""};
            self.value={id: null, name: ""};
            self.idModel = null;
            if (typeof self.onChange === "function") {
                self.onChange(el);
            }
        };;

        self.lookup = function(){
            //console.log("self",self);
            PubSub.publish('lookupModal', {sender:self});
            //alert('test');
            //console.log("lookupCallBack",self.lookupCallBack);
             };


        self.getUrl = function(){
            if (1 == 1 ) {
                if (self.readonly ==true){
                    return "query/get?code=" + self.queryCode + "&perpage=0&page=1&getRowById="+self.idModel;
                }else{
                    return "query/get?code=" + self.queryCode + "&perpage=100&page=1&getRowById="+self.idModel;
                }

            }else{
                return false;
            }
        };;
        self.bind = function(){



            if (self.cacheData && sessionStorage.getItem("componentCache_"+self.getUrl())!=null)
            {

                //console.log("suk emes",sessionStorage.getItem("componentCache_query/get?code="+self.queryCode))
                var data = JSON.parse( sessionStorage.getItem("componentCache_"+self.getUrl()));;
                //self.label = data.title;
                self.values = data.items;
                self.loaded = true;


                if (self.idModel) {


                    var result = $.grep(self.values, function (e) {
                        return e.id == self.idModel;
                    });

                    if (result.length > 0) {
                        self.selectedName = result[0].name;
                        //console.log("result",self.selectedName);
                    }

                    self.model = {id: self.idModel, name: self.selectedName};
                    self.value = self.model;

                }

                return;
            }
            if (self.getUrl()) {
                RestApiService.get(self.getUrl()).
                success(function (data) {

                    self.loaded = true;

                    if (!self.label) {
                        self.label = data.title;
                    }
                    if (self.cacheData) {
                        sessionStorage.setItem("componentCache_" + self.getUrl(), JSON.stringify(data))
                    }

                    self.values = data.items;

                    if (self.idModel) {
                        //var result = $.grep(self.values, function (e) {
                        //    return e.id == self.idModel;
                        //});
                        //
                        //if (result.length > 0) {
                        //    self.selectedName = result[0].name;
                        //    //console.log("result",self.selectedName);
                        //}

                        self.model = data.getSelectedRow;
                        self.value = self.model;
                    }
                });
            }
        };;

        self.refreshSelect = function($select,table){

            //console.log("suka emes",$select);
            if ($select.search.length>=2) {
                RestApiService.get("query/get?code="+table+"&selectContains="+$select.search+"&perpage=100&page=1&getRowById="+self.idModel).
                success(function(data) {
                    console.log(data);
                    $select.items = data.items;
                });

                //RestApiService.get("list/simple/get?code=" + table+"&contains="+$select.search).success(
                //    function (data) {
                //        $select.items = data;
                //    }
                //);
            }

            //$select.items.push({id:-1,title:"vata emes"});
            //return $select;
        };;

        //self.bind();
    }
});


MetronicApp.component('duSelect', {
    template:
    "<div class=\"form-group\">"+
    "                <label ng-if=\"!$ctrl.withoutLabel\" translate>{{$ctrl.label}}</label>"+
    "                <div class=\"input-group col-md-12 col-xs-12\">"+
    "                  <ui-select ng-if=\"$ctrl.withoutSelect!='true' && $ctrl.readonly!='true'\"  style=\"min-width:100px\" ng-change=\"$ctrl.ngModelChange()\" ng-model=\"$ctrl.ngModel\" theme=\"bootstrap\">"+
    "                    <ui-select-match  placeholder=\"{{$ctrl.label | translate}}\">{{$ctrl.rowPrefix}} {{($select.selected.name | translate)}} </ui-select-match>"+
    "                    <ui-select-choices refresh-delay=\"0\" refresh=\"$ctrl.refreshSelect($select,$ctrl.queryCode)\" repeat=\"item in $ctrl.values | filter: $select.search\">"+
    "                      <div style=\"background-color:{{item.background_color}}\" ng-bind-html=\"$ctrl.rowPrefix + ' ' + (item.name | translate) | highlight: $select.search\"></div>"+
    "                    </ui-select-choices>"+
    "                  </ui-select>"+
    "       <label ng-if=\"$ctrl.withoutSelect=='true'\">{{$ctrl.ngModel.name | translate}}</label>"+
    "		<input ng-required autocomplete=\"off\" ng-if=\"$ctrl.withoutSelect!='true' && $ctrl.readonly=='true'\" readonly  ng-value=\"$ctrl.ngModel.name | translate\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+
    "        		  <span ng-if=\"!$ctrl.withoutClear && $ctrl.readonly!='true' && $ctrl.idModel\" title=\"{{'clearCustom' | translate }}\" ng-click=\"$ctrl.clear($ctrl.ngModel)\" class=\"input-group-addon\">"+
    "        		    <i class=\"fa fa-remove\"></i>"+
    "        		  </span>  "+
    "        		  <span ng-if=\"$ctrl.withBrowse\" title=\"{{'browse' | translate }}\" ng-click=\"$ctrl.browse()\" class=\"input-group-addon\">"+
    "        		    <i class=\"fa fa-pencil\"></i>"+
    "        		  </span>  "+
    "        		  <a href=\"{{$ctrl.browseUrl}}\" ng-if=\"$ctrl.browseUrl && $ctrl.idModel \" title=\"{{'BrowseUrlTitle' | translate }}\" class=\"input-group-addon\">"+
    "        		    <i class=\"fa fa-arrow-circle-o-right\"></i>"+
    "        		  </a>  "+
    "        		  <span ng-if=\"$ctrl.lookupPageCode && $ctrl.readonly!='true' \" title=\"{{'OpenLookupPageTitle' | translate }}\" ng-click=\"$ctrl.lookup($ctrl.lookupData)\" class=\"input-group-addon\">"+
    "        		    <i class=\"fa fa-search\"></i>"+
    "        		  </span>  "+
//"        		  <a title=\"{{'Open Entity' | translate }}\"  href=\"#/settings/entitydetails/{{detail._entity_id.id}}\" class=\"input-group-addon\">"+
//"        		    <i class=\"glyphicon glyphicon-chevron-right\"></i>"+
//"        		  </a>                         "+
    "                </div>"+
    "            </div> "+
    "",
    bindings: {
        data: '@',
        label: '@',
        queryCode: '@',
        ngModel: '<',
        value: '=',
        values: '<',
        rowPrefix: '@',
        idModel : '=',
        idLink : '=',
        onChange : '=',
        bindOnInit: '@',
        title: '@',
        withoutLabel: '@',
        withBrowse: '@',
        withoutClear: '@',
        withoutSelect: '@',
        withTranslate: '@',
        cacheData: '@',
        readonly:'@',
        lookupPageCode:'@',
        lookupData:'@',
        lookupCallBack:'<',
        browseUrl:'@',


    },
    require: { ngModelCtrl: 'ngModel' },

    controller: function ($scope,RestApiService,PubSub) {



        var self = this;

        self.loaded = false;

        $scope.$watch(
            "$ctrl.queryCode",
            function handleFooChange( newValue, oldValue ) {
                if (typeof self.queryCode !== "undefined" && self.queryCode !== null  && self.loaded /* && typeof self.idModel !== "undefined" && self.idModel !== null */){
                    console.log("self.queryCode1",self.queryCode);
                    console.log("self.queryCode1",self.idModel);
                    self.bind();

                }


            }
        );

        $scope.$watch(
            "$ctrl.idModel",
            function handleFooChange( newValue, oldValue ) {
                //console.log(oldValue);
                //console.log(newValue);
                //console.log(oldValue);
                if (typeof self.queryCode !== "undefined" && self.queryCode !== null){
                    //self.idOrigModel = new Date(newValue);
                    //console.log("norm777555 "+self.idModel);
                    //console.log("self.queryCode2",self.queryCode)
                    //console.log("self.queryCode2",self.idModel)
                    self.bind();
                }
            }
        );

        self.$onInit = function() {
            //console.log(self.queryCode);
            if (self.bindOnInit) {
                self.bind();
            }
            if (!self.idModel){
                self.idModel = null;
            }

            //console.log("subscribing to bindBuComponents");
            //PubSub.subscribe("bindBuComponents",self.bind);
        };

        self.ngModelChange = function () {
            self.idModel = self.ngModel.id;
            self.ngModelCtrl.$setViewValue(this.ngModel);

        };

        self.edit = function(el){
            console.log("el",el);

            //alert(self.model.id);
            self.idModel = self.ngModel.id;
            self.value = self.ngModel;

            //if (el.id) {
            //    self.idModel = el.id;
            //    self.value = self.idLink;
            //}



            if (typeof self.onChange === "function") {
                el.idLink = self.idLink;
                self.onChange(el);
            }
            //self.onEdit();
        };;

        self.clear = function(el){
            el={id: null, name: ""};
            self.ngModel  =el;
            self.idModel = null;
            //self.$apply();
            self.ngModelCtrl.$setViewValue(el);


        };

        self.lookup = function(lookupData){
            //console.log("self",self);
            //PubSub.publish('lookupModal', {sender:self});

            PubSub.publish("openDetailModal", {callBack:self.lookupCallBack,withoutGo:true, data: lookupData, page:self.lookupPageCode});

        };


        self.getUrl = function(){
            if (1 == 1 ) {
                if (self.readonly ==true){
                    return "query/get?code=" + self.queryCode + "&perpage=0&page=1&getRowById="+self.idModel;
                }else{
                    return "query/get?code=" + self.queryCode + "&perpage=100&page=1&getRowById="+self.idModel;
                }

            }else{
                return false;
            }
        };;
        self.bind = function(){



            if (self.cacheData && sessionStorage.getItem("componentCache_"+self.getUrl())!=null)
            {

                //console.log("suk emes",sessionStorage.getItem("componentCache_query/get?code="+self.queryCode))
                var data = JSON.parse( sessionStorage.getItem("componentCache_"+self.getUrl()));;
                //self.label = data.title;
                self.values = data.items;
                self.loaded = true;


                if (self.idModel) {


                    var result = $.grep(self.values, function (e) {
                        return e.id == self.idModel;
                    });

                    if (result.length > 0) {
                        self.selectedName = result[0].name;
                        //console.log("result",self.selectedName);
                    }

                    self.ngModel = {id: self.idModel, name: self.selectedName};
                    //self.value = self.model;

                }

                return;
            }
            if (self.getUrl()) {
                RestApiService.get(self.getUrl()).
                success(function (data) {

                    self.loaded = true;

                    if (!self.label) {
                        self.label = data.title;
                    }
                    if (self.cacheData) {
                        sessionStorage.setItem("componentCache_" + self.getUrl(), JSON.stringify(data))
                    }

                    self.values = data.items;

                    if (self.idModel) {
                        //var result = $.grep(self.values, function (e) {
                        //    return e.id == self.idModel;
                        //});
                        //
                        //if (result.length > 0) {
                        //    self.selectedName = result[0].name;
                        //    //console.log("result",self.selectedName);
                        //}

                        self.ngModel = data.getSelectedRow;
                        //self.value = self.ngModel;
                    }
                });
            }
        };;

        self.refreshSelect = function($select,table){

            //console.log("suka emes",$select);
            if ($select.search.length>=2) {
                RestApiService.get("query/get?code="+table+"&selectContains="+$select.search+"&perpage=100&page=1&getRowById="+self.idModel).
                success(function(data) {
                    console.log(data);
                    $select.items = data.items;
                });

                //RestApiService.get("list/simple/get?code=" + table+"&contains="+$select.search).success(
                //    function (data) {
                //        $select.items = data;
                //    }
                //);
            }

            //$select.items.push({id:-1,title:"vata emes"});
            //return $select;
        };;

        //self.bind();
    }
});


MetronicApp.component('buCheckbox', {
    template: "" +
    "<div class=\"form-group\" >" +
    "<label><checkbox ng-if=\"  ($ctrl.readonly != 'true')   \" ng-model=\"$ctrl.idModel\" ng-click=\"$ctrl.edit($ctrl.idModel)\" ng-true-value=\"1\"   ng-false-value=\"0\"  ></checkbox>\n" +
    "<i class='fa fa-check-square-o'  ng-if=\"  ($ctrl.readonly == 'true') && $ctrl.idModel == 1  \" ></i> " +
    "<i class='fa fa-square-o'  ng-if=\"  ($ctrl.readonly == 'true') && $ctrl.idModel != 1  \" ></i> " +
    "<translate>{{$ctrl.label}}</translate></label>" +
    "</div>",
    bindings: {
        label: '@',
        idModel : '=',
        onChange : '=',
        onClick : '=',
        readonly: '@',
        title: '@'
    },

    controller: function () {
        var self = this;

        self.click = function(data){

            if (typeof self.onClick === "function") {
                self.onClick(data);
            }
        }

        self.edit = function(data){

            if (typeof self.onChange === "function") {
                self.onChange(data);
            }
        }
    }
});



MetronicApp.component('buPasswordLabel', {
    template: "" +
    "<div class=\"form-group\">"+
    "	<label translate>{{$ctrl.label}}</label>"+
    "	<div class=\"input-group\">"+
    "		<span class=\"input-group-addon\">"+
    "		<i class=\"{{$ctrl.icon}}\"></i>"+
    "		</span>"+
    "		<input ng-required ng-change=\"$ctrl.edit()\" type=\"password\"  ng-model=\"$ctrl.idModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+
    "	</div>"+
    "</div>",
    bindings: {
        label: '@',
        idModel : '=',
        onChange : '=',
        title: '@',
        icon: '@',
        type: '@'
    },
    controller: function () {
        var self = this;

        //self.Model = ""

        self.edit = function(){

            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }

            //console.log("suk emes2226776  "+self.idModel);
        }

    }
});

MetronicApp.component('buInputLabel', {
    template: "" +
    "<div class=\"form-group\">"+
    "	<label translate>{{$ctrl.label}}</label>"+
    "	<div class=\"input-group\" ng-if=\"$ctrl.icon\">"+
    "		<span ng-if=\"!$ctrl.withoutIcon\"  class=\"input-group-addon\">"+
    "		<i class=\"{{$ctrl.icon}}\"></i>"+
    "		</span>"+
    "		<input ng-required ng-readonly=\" ($ctrl.readonly == 'true') \"  ng-change=\"$ctrl.edit()\"  ng-model=\"$ctrl.idModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.placeholder?$ctrl.placeholder:$ctrl.label | translate }}\" >"+
    "	</div>"+
    "		<input ng-if=\"!$ctrl.icon\" ng-required ng-readonly=\" ($ctrl.readonly == 'true') \"  ng-change=\"$ctrl.edit()\"  ng-model=\"$ctrl.idModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.placeholder?$ctrl.placeholder:$ctrl.label | translate }}\" >"+
    "</div>",
    bindings: {
        label: '@',
        withoutIcon: '@',
        idModel : '=',
        onChange : '=',
        title: '@',
        placeholder: '@',
        icon: '@',
        type: '@',
        readonly: '@'
    },
    controller: function () {
        var self = this;

        //self.Model = ""

        self.edit = function(){

            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }

            //console.log("suk emes2226776  "+self.idModel);
        }

    }
});

MetronicApp.component('buInputPassword', {
    template: "" +
    "<div class=\"form-group\">"+
    "	<label translate>{{$ctrl.label}}</label>"+
    "	<div class=\"input-group\">"+
    "		<span ng-if=\"!$ctrl.withoutIcon\"  class=\"input-group-addon\">"+
    "		<i class=\"{{$ctrl.icon}}\"></i>"+
    "		</span>"+
    "		<input ng-required type=\"password\" ng-readonly=\" ($ctrl.readonly == 'true') \"  ng-change=\"$ctrl.edit()\"  ng-model=\"$ctrl.idModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+
    "	</div>"+
    "</div>",
    bindings: {
        label: '@',
        withoutIcon: '@',
        idModel : '=',
        onChange : '=',
        title: '@',
        icon: '@',
        type: '@',
        readonly: '@'
    },
    controller: function () {
        var self = this;

        //self.Model = ""

        self.edit = function(){

            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }

            //console.log("suk emes2226776  "+self.idModel);
        }

    }
});

MetronicApp.component('duInputNumberLabel', {
    template: "" +
    "<div class=\"form-group\">"+
    "	<label ng-if=\"!$ctrl.withoutLabel\" translate>{{$ctrl.label}}</label>"+
    "	<div class=\"input-group\">"+
    "		<span class=\"input-group-addon\">"+
    "		<i class=\"{{$ctrl.icon}}\"></i>"+
    "		</span>"+
    "		<input style=\" {{ $ctrl.inputStyle }}\" ng-required ng-change=\"$ctrl.ngModelChange()\" ng-readonly=\" ($ctrl.readonly == 'true') \" type=\"number\" ng-model=\"$ctrl.ngModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+
    "	</div>"+
    "</div>",
    bindings: {
        label: '@',
        withoutLabel: '@',
        inputStyle: '@',
        title: '@',
        readonly: '@',
        icon: '@',
        type: '@',
        ngModel: '<'
    },
    require: { ngModelCtrl: 'ngModel' },

    controller: function (PubSub,$scope) {
        var self = this;
        $scope.$watch(
            "$ctrl.ngModel",
            function handleFooChange( newValue, oldValue ) {
                //console.log("newvalue",newValue);
                if (newValue !== undefined){
                    if (typeof newValue == "string") {
                        self.ngModel = newValue*1;
                    }
                }
            }
        );
        self.ngModelChange = function () {
            self.ngModelCtrl.$setViewValue(self.ngModel);
        };

        self.bind = function(){
            console.log("newvalue",self.idModel);
            self.idModel = self.idModel * 1;
        };;

        self.$onInit = function() {
            if (self.idModel) {
                self.idModel = self.idModel * 1;
            }else{
                self.idModel = 0;
            }
            //PubSub.subscribe("bindBuComponents", self.bind);
        };;

        self.edit = function(){

            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }
            //console.log("suk emes2226776  "+self.idModel);
        }

    }
});


MetronicApp.component('buInputNumberLabel', {
    template: "" +
    "<div class=\"form-group\">"+
    "	<label ng-if=\"!$ctrl.withoutLabel\" translate>{{$ctrl.label}}</label>"+
    "	<div ng-if=\"$ctrl.icon\" class=\"input-group\">"+
    "		<span class=\"input-group-addon\">"+
    "		<i class=\"{{$ctrl.icon}}\"></i>"+
    "		</span>"+
    "		<input style=\" {{ $ctrl.inputStyle }}\" ng-required ng-change=\"$ctrl.edit()\" ng-readonly=\" ($ctrl.readonly == 'true') \" type=\"number\" ng-model=\"$ctrl.idModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+
    "	</div>"+
    "<input ng-if=\"!$ctrl.icon\"  style=\" {{ $ctrl.inputStyle }}\" ng-required ng-change=\"$ctrl.edit()\" ng-readonly=\" ($ctrl.readonly == 'true') \" type=\"number\" ng-model=\"$ctrl.idModel\"  class=\"form-control\" placeholder=\"{{ $ctrl.label | translate }}\" >"+

    "</div>",
    bindings: {
        label: '@',
        withoutLabel: '@',
        inputStyle: '@',
        idModel : '=',
        onChange : '=',
        title: '@',
        readonly: '@',
        icon: '@',
        type: '@'
    },
    controller: function (PubSub,$scope) {
        var self = this;

        //self.Model = ""

        $scope.$watch(
            "$ctrl.idModel",
            function handleFooChange( newValue, oldValue ) {
                //console.log("newvalue",newValue);
                if (newValue !== undefined){
                    if (typeof newValue == "string") {
                        self.idModel = newValue*1;
                    }
                }
            }
        );

        self.bind = function(){
            console.log("newvalue",self.idModel);
            self.idModel = self.idModel * 1;
        };;

        self.$onInit = function() {
            if (self.idModel) {
                self.idModel = self.idModel * 1;
            }else{
                self.idModel = 0;
            }
            //PubSub.subscribe("bindBuComponents", self.bind);
        };;

        self.edit = function(){

            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }
            //console.log("suk emes2226776  "+self.idModel);
        }

    }
});


MetronicApp.component('buDateTime', {
    template: "" +
    "<div class=\"form-group\">"+
    "	<label class=\"control-label\" translate>{{$ctrl.label}}</label>"+
    "		<div class=\"input-group\">"+
        "			<span class=\"input-group-addon\">"+
    "				<i class=\"{{$ctrl.icon}}\"></i>"+
        "			</span>"+
    "			<input ng-readonly=\" ($ctrl.readonly == 'true') \" min=\"1900-01-01T00:00:00\" max=\"2070-12-31T00:00:00\" type=\"datetime-local\"  step=1 ng-change=\"$ctrl.edit($ctrl.idOrigModel)\" aria-describedby=\"deliverytime-addon\" "+
        "			   ng-model= \"$ctrl.idOrigModel\"  class=\"form-control\" placeholder=\"{{$ctrl.label}}\" >"+
    "		</div>"+
    "</div>  "
    ,
    bindings: {
        label: '@',
        idModel : '=',
        onChange : '=',
        readonly: '@',
        title: '@',
        icon: '@'
    },
    controller: function ($scope) {
        var self = this;

        $scope.$watch(
            "$ctrl.idModel",
            function handleFooChange( newValue, oldValue ) {

                console.log("datetime",newValue);

                if ( newValue !== undefined){
                    if (typeof newValue == "string") {
                        var t = newValue.split(/[- :]/);
                        if (t && t.length == 3) {
                            t[3] = 0;
                            t[4] = 0;
                            t[5] = 0;
                        }
                        self.idOrigModel = new Date(t[0], t[1] - 1, t[2], t[3], t[4], t[5]);
                        //self.idOrigModel = new Date(newValue);
                        //console.log(newValue.replace(" ", "T"),"date");
                    }else if (typeof newValue == "object"){
                        self.idOrigModel = newValue;
                    }
                }
            }
        );

        self.edit = function(el){

            var date = new Date(el);
            var day = date.getDate();
            var month = date.getMonth();
            var year = date.getFullYear();
            var hour = date.getHours();
            var minutes = date.getMinutes();
            var seconds = date.getSeconds();

            month++;;
            if (month<10){
                month = "0"+month;
            }
            if (day<10){
                day = "0"+day;
            }
            if (hour<10){
                hour = "0"+hour;
            }
            if (minutes<10){
                minutes = "0"+minutes;
            }
            if (seconds<10){
                seconds = "0"+seconds;
            }

            self.idModel = year + "-" + month + "-" + day+" "+hour+":"+minutes+":"+seconds;
            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }


        };;

        self.$onInit = function() {
            //console.log("self.icon",self.icon.length);
            if (!self.icon) {
                self.icon = "fa fa-clock-o";
            }

            //console.log("suk1"+self.idModel);
            //self.idOrigModel = new Date(self.idModel);
            //console.log("suk2"+self.idOrigModel);
        }
    }
});

MetronicApp.component('buDate', {
    //template: "" +
    //"<div class=\"form-group\">"+
    //"	<label class=\"control-label\" translate>{{$ctrl.label}}</label>"+
    //"		<div class=\"input-group\">"+
    //"			<span class=\"input-group-addon\">"+
    //"				<i class=\"{{$ctrl.icon}}\"></i>"+
    //"			</span>"+
    //"			<input ng-readonly=\" ($ctrl.readonly == 'true') \" min=\"1900-01-01T00:00:00\" max=\"2070-12-31T00:00:00\" type=\"date\"  step=1 ng-change=\"$ctrl.edit($ctrl.idOrigModel)\" aria-describedby=\"deliverytime-addon\" "+
    //"			   ng-model= \"$ctrl.idOrigModel\"  class=\"form-control\" placeholder=\"{{$ctrl.label}}\" >"+
    //"		</div>"+
    //"</div>  "
    //,

    template: "" +
    "<div class=\"form-group\">"+
    "	<label ng-if=\"$ctrl.label\" class=\"control-label\" translate>{{$ctrl.label}}</label>"+
    "		<div class=\"input-group\">"+
    "			<input uib-datepicker-popup=\"{{$ctrl.format}}\" is-open=\"$ctrl.opened\" datepicker-options=\"$ctrl.dateOptions\"" +
    "           placeholder=\"ДД.ММ.ГГГГ\" ng-required=\"true\" close-text=\"Close\" alt-input-formats=\"$ctrl.altInputFormats\" " +
    "            ng-readonly=\" ($ctrl.readonly == 'true') \"  type=\"text\"  ng-change=\"$ctrl.edit($ctrl.idOrigModel)\"" +
    "			 ng-model= \"$ctrl.idOrigModel\"  class=\"form-control\"  >"+
    "<span class=\"input-group-btn\">"+
    "<button type=\"button\" class=\"btn btn-default\" ng-click=\"$ctrl.open()\"><i class=\"glyphicon glyphicon-calendar\"></i></button>"+
    "</span>"+
    "		</div>"+

    "</div>  ",
    bindings: {
        label: '@',
        idModel : '=',
        onChange : '=',
        readonly: '@',
        title: '@',
        icon: '@'
    },
    controller: function ($scope) {
        var self = this;

        $scope.$watch(
            "$ctrl.idModel",
            function handleFooChange( newValue, oldValue ) {

                //console.log(newValue,"newValue");

                if (newValue !== undefined){
                    //var t =newValue.split(/[-]/);
                    //self.idOrigModel = new Date(t[0], t[1]-1, t[2]);



                    if (typeof newValue == "string") {

                        var t = newValue.substring(0,10).split(/[-]/);
                        //console.log("test2",new Date(t[0], t[1] - 1, t[2],0,0,0));
                        //console.log("test2",t);
                        self.idOrigModel = new Date(t[0], t[1] - 1, t[2],0,0,0);
                    }else if (typeof newValue == "object"){
                        //console.log("test");
                        self.idOrigModel = newValue;
                    }

                }
            }
        );

        self.open = function() {
            self.opened = true;
        };

        self.opened = false;

        self.formats = ['dd-MMMM-yyyy', 'yyyy/MM/dd', 'dd.MM.yyyy', 'shortDate'];
        self.format = self.formats[2];
        self.altInputFormats = ['M!/d!/yyyy'];

        self.dateOptions = {
            dateDisabled: self.disabled,
            formatYear: 'yy',
            maxDate: new Date(2070, 5, 22),
            minDate: new Date(1900,1,1),
            startingDay: 1
        };

        function disabled(data) {
            return true;
        }
        function disabled2(data) {
            var date = data.date,
                mode = data.mode;
            return mode === 'day' && (date.getDay() === 0 || date.getDay() === 6);
        }

        self.edit = function(el){

            //console.log("test");
            var date = new Date(el);

            var day = date.getDate();
            var month = date.getMonth();
            var year = date.getFullYear();
            var hour = date.getHours();
            var minutes = date.getMinutes();
            var seconds = date.getSeconds();

            month++;;
            if (month<10){
                month = "0"+month;
            }
            if (day<10){
                day = "0"+day;
            }
            if (hour<10){
                hour = "0"+hour;
            }
            if (minutes<10){
                minutes = "0"+minutes;
            }
            if (seconds<10){
                seconds = "0"+seconds;
            }

            self.idModel = year + "-" + month + "-" + day+" "+hour+":"+minutes+":"+seconds;
            if (typeof self.onChange === "function") {
                self.onChange(self.idModel);
            }


        };;

        self.$onInit = function() {
            if (!self.icon) {
                self.icon = "fa fa-clock-o";
            }
            //console.log("suk1"+self.idModel);
            //self.idOrigModel = new Date(self.idModel);
            //console.log("suk2"+self.idOrigModel);
        }
    }
});


MetronicApp.component('duDate', {


    template: "" +
    "<div class=\"form-group\">"+

    "	<label ng-if=\"$ctrl.label\" class=\"control-label\" translate>{{$ctrl.label}}</label>"+
    "		<div class=\"input-group\">"+
    "			<input ng-change=\"$ctrl.ngModelChange()\" ng-model=\"$ctrl.dateModel\" uib-datepicker-popup=\"{{$ctrl.format}}\" is-open=\"$ctrl.opened\" datepicker-options=\"$ctrl.dateOptions\"" +
    "           placeholder=\"{{'DD.MM.YYYY' | translate}}\" close-text=\"Close\" alt-input-formats=\"$ctrl.altInputFormats\" " +
    "            ng-readonly=\" ($ctrl.readonly == 'true') \"" +
    "			  class=\"form-control\"  >"+
    "<span class=\"input-group-btn\">"+
    "<button type=\"button\" class=\"btn btn-default\" ng-click=\"$ctrl.open()\"><i class=\"glyphicon glyphicon-calendar\"></i></button>"+
    "</span>"+
    "</div>"+
    "</div>  ",
    bindings: {
        label: '@',
        ngModel : '<',
        onChange : '=',
        readonly: '@',
        title: '@',
        icon: '@'
    },
    require: { ngModelCtrl: 'ngModel' },
    controller: function ($scope,moment) {
        var self = this;

        $scope.$watch(
            "$ctrl.ngModel",
            function handleFooChange( newValue, oldValue ) {
                if (newValue !== undefined){
                    if (typeof newValue == "string") {
                        var date = moment(newValue,"YYYY-MM-DD HH:mm");
                        self.dateModel = new Date(date);
                        self.hours = self.dateModel.getHours();
                        self.minutes= self.dateModel.getMinutes();
                        //console.log(date);
                    }
                }
            }
        );

        self.ngModelChange = function() {
            //self.dateModel.setMinutes(self.minutes);
            //self.dateModel.setHours(self.hours);
            //console.log("datatype",typeof self.dateModel);
            //console.log("value",self.dateModel);
            if (self.dateModel == null) {
                self.ngModel =null;
                self.ngModelCtrl.$setViewValue(self.ngModel);
            }else {
                date = moment(self.dateModel);
                self.ngModel = date.format("YYYY-MM-DD");
                self.ngModelCtrl.$setViewValue(self.ngModel);
            }
        };

        self.setHour = function (h){

            if (typeof h == "undefined"){
                return
            }
            if (h>23 || h <0){
                h = 0;
            }
            console.log("h",h);
            self.dateModel.setHours(h);
            date = moment(self.dateModel);
            self.ngModel = date.format("YYYY-MM-DD");
            self.ngModelCtrl.$setViewValue(self.ngModel);
        }
        self.setMinute = function (m){

            if (typeof m == "undefined"){
                return
            }

            if (m>59 || m <0){
                m = 0;
            }
            self.dateModel.setMinutes(m);
            date = moment(self.dateModel);
            self.ngModel = date.format("YYYY-MM-DD");
            self.ngModelCtrl.$setViewValue(self.ngModel);
        }

        self.open = function() {
            self.opened = true;
        };

        self.opened = false;
        self.minutes = 0;
        self.hours = 0;
        self.seconds = 0;

        self.formats = ['dd-MMMM-yyyy', 'yyyy/MM/dd', 'dd.MM.yyyy', 'shortDate'];
        self.format = self.formats[2];
        self.altInputFormats = ['M!/d!/yyyy'];

        self.dateOptions = {
            dateDisabled: self.disabled,
            formatYear: 'yy',
            maxDate: new Date(2070, 5, 22),
            minDate: new Date(1900,1,1),
            startingDay: 1
        };

        self.$onInit = function() {
            if (!self.icon) {
                self.icon = "fa fa-clock-o";
            }
        }
    }
});


MetronicApp.component('duDateTime', {
    //template: "" +
    //"<div class=\"form-group\">"+
    //"	<label class=\"control-label\" translate>{{$ctrl.label}}</label>"+
    //"		<div class=\"input-group\">"+
    //"			<span class=\"input-group-addon\">"+
    //"				<i class=\"{{$ctrl.icon}}\"></i>"+
    //"			</span>"+
    //"			<input ng-readonly=\" ($ctrl.readonly == 'true') \" min=\"1900-01-01T00:00:00\" max=\"2070-12-31T00:00:00\" type=\"date\"  step=1 ng-change=\"$ctrl.edit($ctrl.idOrigModel)\" aria-describedby=\"deliverytime-addon\" "+
    //"			   ng-model= \"$ctrl.idOrigModel\"  class=\"form-control\" placeholder=\"{{$ctrl.label}}\" >"+
    //"		</div>"+
    //"</div>  "
    //,

    template: "" +
    "<div class=\"form-group\">"+

    "	<label ng-if=\"$ctrl.label\" class=\"control-label\" translate>{{$ctrl.label}}</label>"+
    "		<div class=\"input-group\">"+
    "			<input ng-change=\"$ctrl.ngModelChange()\" ng-model=\"$ctrl.dateModel\" uib-datepicker-popup=\"{{$ctrl.format}}\" is-open=\"$ctrl.opened\" datepicker-options=\"$ctrl.dateOptions\"" +
    "           placeholder=\"{{'DD.MM.YYYY' | translate}}\" close-text=\"Close\" alt-input-formats=\"$ctrl.altInputFormats\" " +
    "            ng-readonly=\" ($ctrl.readonly == 'true') \"" +
    "			  class=\"form-control\"  >"+
    "<span class=\"input-group-btn\">"+
    "<button type=\"button\" class=\"btn btn-default\" ng-click=\"$ctrl.open()\"><i class=\"glyphicon glyphicon-calendar\"></i></button>"+
    "</span>"+
    "<span class=\"input-group-btn\">"+
    "<input ng-model= \"$ctrl.hours\" ng-change=\"$ctrl.setHour($ctrl.hours)\"  type=\"number\" placeholder=\"{{'HH' | translate}}\" style=\"min-width:40px\" class=\"form-control\"/> "+
    "</span>"+
    "<span class=\"input-group-btn\">"+
    "<label>:</label>"+
    "</span>"+
    "<span class=\"input-group-btn\">"+
    "<input ng-model= \"$ctrl.minutes\" min=0 max=60 step=\"5\" ng-change=\"$ctrl.setMinute($ctrl.minutes)\" type=\"number\" placeholder=\"{{'MM' | translate}}\" style=\"min-width:40px\" class=\"form-control\"/> "+
    "</span>"+
    "<span class=\"input-group-btn\">"+
    "<label>:</label>"+
    "</span>"+
    "<span class=\"input-group-btn\">"+
    "<input ng-model= \"$ctrl.seconds\" min=0 max=60 step=\"10\" ng-change=\"$ctrl.setSecond($ctrl.seconds)\" type=\"number\" placeholder=\"{{'СС' | translate}}\" style=\"min-width:40px\" class=\"form-control\"/> "+
    "</span>"+


    "</div>"+
    "</div>  ",
    bindings: {
        label: '@',
        ngModel : '<',
        onChange : '=',
        readonly: '@',
        title: '@',
        minuteStep: '@',
        icon: '@'
    },
    require: { ngModelCtrl: 'ngModel' },
    controller: function ($scope,moment) {
        var self = this;

        $scope.$watch(
            "$ctrl.ngModel",
            function handleFooChange( newValue, oldValue ) {
                if (newValue !== undefined){
                    if (typeof newValue == "string") {
							console.log("z1",date);
							console.log("z2",newValue);
                            var date = moment(newValue,"YYYY-MM-DD HH:mm:ss");
                            self.dateModel = new Date(date);
                            self.hours = self.dateModel.getHours();
                            self.minutes= self.dateModel.getMinutes();
							self.seconds= self.dateModel.getSeconds();

                            console.log(date);
                    }
                }
            }
        );

        self.ngModelChange = function() {
			self.dateModel.setSeconds(self.seconds);
            self.dateModel.setMinutes(self.minutes);
            self.dateModel.setHours(self.hours);
            date = moment(self.dateModel);
            self.ngModel = date.format("YYYY-MM-DD HH:mm:ss");
            self.ngModelCtrl.$setViewValue(self.ngModel);
        };

        self.setHour = function (h){

            if (typeof h == "undefined"){
                return
            }
            if (h>23 || h <0){
                h = 0;
            }
            console.log("h",h);
            self.dateModel.setHours(h);
            date = moment(self.dateModel);
            self.ngModel = date.format("YYYY-MM-DD HH:mm:ss");
            self.ngModelCtrl.$setViewValue(self.ngModel);
        }

        self.setSecond = function (s){

            if (typeof s == "undefined"){
                return
            }

            if (s>59 || s <0){
                s = 0;
            }
            self.dateModel.setSeconds(s);
            date = moment(self.dateModel);
            self.ngModel = date.format("YYYY-MM-DD HH:mm:ss");
            self.ngModelCtrl.$setViewValue(self.ngModel);
        }

        self.setMinute = function (m){

            if (typeof m == "undefined"){
                return
            }

            if (m>59 || m <0){
                m = 0;
            }
            self.dateModel.setMinutes(m);
            date = moment(self.dateModel);
            self.ngModel = date.format("YYYY-MM-DD HH:mm:ss");
            self.ngModelCtrl.$setViewValue(self.ngModel);
        }

        self.open = function() {
            self.opened = true;
        };

        self.opened = false;
        self.minutes = 0;
        self.hours = 0;
        self.seconds = 0;

        self.formats = ['dd-MMMM-yyyy', 'yyyy/MM/dd', 'dd.MM.yyyy', 'shortDate'];
        self.format = self.formats[2];
        self.altInputFormats = ['M!/d!/yyyy'];

        self.dateOptions = {
            dateDisabled: self.disabled,
            formatYear: 'yy',
            maxDate: new Date(2070, 5, 22),
            minDate: new Date(1900,1,1),
            startingDay: 1
        };

        self.$onInit = function() {
            if (!self.icon) {
                self.icon = "fa fa-clock-o";
            }
        }
    }
});


MetronicApp.component('buImageUploaderByUrl', {
    template: "" +
    "<a ng-click=\"$ctrl.changeImage(this)\" class=\"thumbnail\">"+
    "<img src=\"{{$ctrl.imgUrl}}\" alt=\"...\">"+
    "<input id=\"{{$ctrl.fileInputElementId}}\" style=\"display:none\" type=\"file\" nv-file-select uploader=\"$ctrl.uploader\"/><br/>"+
    "</a>",
    bindings: {
        dir:"@",
        label: '@',
        url : '=',
        onChange : '=',
        title: '@',
        icon: '@',
        fileUuid: '@'
    },
    controller: function ($scope,FileUploader) {



        var self = this;

        $scope.$watch(
            "$ctrl.url",
            function handleFooChange( newValue, oldValue ) {
                if (newValue){
                    console.log("norm777 "+newValue+" "+self.url);
                    self.imgUrl=self.url;
                }
            }
        );

        //self.Model = ""

        self.$onInit  = function(){
            var uniqid = Date.now();
            self.fileInputElementId = "fileuploader"+uniqid;
        };;

        self.changeImage = function(x){
            document.getElementById(self.fileInputElementId).click();
        };;

        if (!self.dir){
            self.dir = "PROF"
        }
        self.uploader = new FileUploader({ autoUpload:true, url: '../restapi/upload?dir='+self.dir });

        self.uploader.onCompleteItem = function(fileItem, response, status, headers) {
            fileItem.result=response;
            self.url = response.url;
            if (typeof self.onChange === "function") {
                self.onChange();
            }
        };


        self.edit = function(){
            if (typeof self.onChange === "function") {
                self.onChange();
            }
        }

    }
});

MetronicApp.component('buImageUploader', {
    template: "" +
    "<a ng-click=\"$ctrl.changeImage(this)\" class=\"thumbnail\">"+
        "<img src=\"{{$ctrl.imgUrl}}\" alt=\"...\">"+
        "<input id=\"{{$ctrl.fileInputElementId}}\" style=\"display:none\" type=\"file\" nv-file-select uploader=\"$ctrl.uploader\"/><br/>"+
    "</a>",
    bindings: {
        label: '@',
        imgUrl: '=',
        idModel : '=',
        onChange : '=',
        title: '@',
        icon: '@',
        fileUuid: '@'
    },
    controller: function ($scope,FileUploader,$timeout) {



        var self = this;

        $scope.$watch(
            "$ctrl.fileUuid",
            function handleFooChange( newValue, oldValue ) {
                if (newValue){
                    console.log("norm777 "+newValue+" "+self.fileUuid);
                    self.imgUrl="/restapi/getfile?code="+self.fileUuid;
                }
            }
        );

        //self.Model = ""

        self.$onInit  = function(){
            var uniqid = Date.now();
            self.fileInputElementId = "fileuploader"+uniqid;
        };;

        //self.changeImage = function() {
        //    var currentButton = angular.element(document.getElementById(self.fileInputElementId));
        //    $timeout(function () {
        //        currentButton.triggerHandler("click");
        //    });
        //}

        self.changeImage = function(x){
            document.getElementById(self.fileInputElementId).click();
            //angular.element('#'+self.fileInputElementId).trigger('click');
        };;
        self.uploader = new FileUploader({ autoUpload:true, url: '../restapi/upload?dir=PROF' });

        self.uploader.onCompleteItem = function(fileItem, response, status, headers) {
            fileItem.result=response;
            self.imgUrl="/restapi/getfile?code="+response.guid;
            self.idModel =response.id;
            if (typeof self.onChange === "function") {
                self.onChange();
            }
        };


        self.edit = function(){
            if (typeof self.onChange === "function") {
                self.onChange();
            }
        }

    }
});

MetronicApp.component('buFileUploaderByUrl', {
    template: "" +
    "<input id=\"{{$ctrl.fileInputElementId}}\" type=\"file\" nv-file-select uploader=\"$ctrl.uploader\"/><br/>",

    bindings: {
        label: '@',
        onChange : '=',
        url : '=',
        onBeforeUploadItem: '=',
        title: '@',
        dir: '@',
        icon: '@'

    },
    controller: function ($scope,FileUploader,$timeout) {


        var self = this;



        self.bind = function(){

            if (!self.dir){
                self.dir = "PROF"
            }
            self.uploader = new FileUploader({ autoUpload:true, url: '../restapi/upload?dir='+self.dir });

            self.uploader.onBeforeUploadItem = function (fileItem){
                if (typeof self.onBeforeUploadItem === "function") {
                    self.onBeforeUploadItem(fileItem);
                }

            };
            self.uploader.onCompleteItem = function(fileItem, response, status, headers) {
                //fileItem.result=response;
                //self.uuidModel=response.guid;
                self.url=response.url;
                //self.restServiceOutput = response.restServiceOutput;
                //self.idModel =response.id;
                if (typeof self.onChange === "function") {
                    self.onChange(response);
                }
            };
        };;

        self.$onInit  = function(){
            var uniqid = Date.now();
            self.fileInputElementId = "fileuploader"+uniqid;
            self.bind();
        }







    }
});


MetronicApp.component('buFileUploaderByData', {
    template: "" +
    "<input id=\"{{$ctrl.fileInputElementId}}\" type=\"file\" nv-file-select uploader=\"$ctrl.uploader\"/><br/>",

    bindings: {
        label: '@',
        onChange : '=',
        data : '=',
        onBeforeUploadItem: '=',
        title: '@',
        dir: '@',
        icon: '@'

    },
    controller: function ($scope,FileUploader,$timeout) {


        var self = this;



        self.bind = function(){

            if (!self.dir){
                self.dir = "PROF"
            }
            self.uploader = new FileUploader({ autoUpload:true, url: '../restapi/upload?dir='+self.dir });

            self.uploader.onBeforeUploadItem = function (fileItem){
                if (typeof self.onBeforeUploadItem === "function") {
                    self.onBeforeUploadItem(fileItem);
                }

            };
            self.uploader.onCompleteItem = function(fileItem, response, status, headers) {
                //fileItem.result=response;
                //self.uuidModel=response.guid;
                self.data=response;
                //self.restServiceOutput = response.restServiceOutput;
                //self.idModel =response.id;
                if (typeof self.onChange === "function") {
                    self.onChange(response);
                }
            };
        };;

        self.$onInit  = function(){
            var uniqid = Date.now();
            self.fileInputElementId = "fileuploader"+uniqid;
            self.bind();
        }







    }
});


MetronicApp.component('buFileUploader', {
    template: "" +
    "<input id=\"{{$ctrl.fileInputElementId}}\" type=\"file\" nv-file-select uploader=\"$ctrl.uploader\"/><br/>",

    bindings: {
        label: '@',
        restServiceOutput : '=',
        idModel : '=',
        uuidModel : '=',
        onChange : '=',
        onBeforeUploadItem: '=',
        title: '@',
        dir: '@',
        icon: '@'

    },
    controller: function ($scope,FileUploader,$timeout) {


        var self = this;



        self.bind = function(){

            if (!self.dir){
                self.dir = "PROF"
            }
            self.uploader = new FileUploader({ autoUpload:true, url: '../restapi/upload?dir='+self.dir });

            self.uploader.onBeforeUploadItem = function (fileItem){
                if (typeof self.onBeforeUploadItem === "function") {
                    self.onBeforeUploadItem(fileItem);
                }

            };;
            self.uploader.onCompleteItem = function(fileItem, response, status, headers) {
                //fileItem.result=response;
                self.uuidModel=response.guid;
                self.restServiceOutput = response.restServiceOutput;
                self.idModel =response.id;
                if (typeof self.onChange === "function") {
                    self.onChange(response);
                }
            };
        };;

        self.$onInit  = function(){
            var uniqid = Date.now();
            self.fileInputElementId = "fileuploader"+uniqid;
            self.bind();
        }







    }
});


MetronicApp.component('buCheckboxList', {
    template: "<div class=\"form-group\">"+
    "                <label translate>{{$ctrl.label}}</label>"+
        "<div class=\"input-group\">"+
    "                <div ng-repeat='item in $ctrl.values' class=\"{{$ctrl.checkboxClass}}\">"+
                        "<bu-checkbox  ng-click=\"$ctrl.click()\" label='{{item.title}}' id-model='item.enable'></bu-checkbox>"+
    "   </div>"+
    "                </div>"+
    "            </div> ",
    bindings: {
        data: '@',
        label: '@',
        queryCode: '@',
        value: '@',
        masterId: '@',
        masterField: '@',
        foreignField: '@',
        idModel : '=',
        title: '@',
        onChange: '=',
        getDmlArray: '=',
        refresh: '=',
        checkboxClass: '@'

    },
    controller: function ($scope,RestApiService,PubSub) {



        var self = this;


        self.dmlArray = [];
        if (!self.checkboxClass){
            self.checkboxClass = "col-md-4"
        }

        self.click = function(){

            if (typeof self.onChange === "function") {
                self.onChange();
            }
        };;

        self.refresh = function(){
            self.bind();
            //console.log(self.dmlArray,"self.dmlArray");
        };;

        self.getDmlArray = function(){
            return self.calcDmlArray();
            //console.log(self.dmlArray,"self.dmlArray");
        };;

        self.calcDmlArray = function(){
            //self.dmlArray = [1,2,3];

            self.dmlInsertArray = self.values.filter(function(item) {
                return (item.enable == 1)
            });

            self.dmlInsertArray = self.dmlInsertArray.map(function(item) {
                i = {};;
                i[self.masterField] = self.masterId;
                i[self.foreignField] = item[self.foreignField];
                return i
            });

            self.dmlDeleteArray = self.values.filter(function(item) {
                return item["id"] != null;
            });

            self.dmlDeleteArray = self.dmlDeleteArray.map(function(item) {
                i = {};;
                i["id"] = item.id;
                return i
            });


            self.dmlArray = [
                {table_name:self.entityCode,action:"delete",values:  self.dmlDeleteArray},
                {table_name:self.entityCode,action:"insert",values:  self.dmlInsertArray}];;

            return self.dmlArray;

        };;

        self.$onInit = function() {
            //console.log(self.queryCode);
            self.bind();
            PubSub.subscribe("bindBuComponents",self.bind);
        };

        self.edit = function(){
                console.log("edited");
        };;

        self.clear = function(){
            self.model={id: null, name: ""};
        };;

        self.bind = function(){

            RestApiService.get("query/get?code="+self.queryCode+"&param1="+self.masterId).
            success(function(data) {
                if(!self.label){
                    self.label = data.title;
                }
                self.values = data.items;
                self.entityCode = data.entityCode;


                var result = $.grep(self.values, function(e){ return e.id == self.idModel; });

                if ( result.length > 0 )
                {
                    self.selectedName = result[0].name;
                    //console.log("result",self.selectedName);
                }

                self.model={id: self.idModel,name:self.selectedName};
            });
        }
    }
});


MetronicApp.component('buQueryBinder', {
    template: "" +
    "",
    bindings: {
        queryCode: '@',
        type: '@',
        bindOnInit : '@',
        values: '=',
    },
    controller: function (RestApiService,PubSub) {
        var self = this;

        console.log("buQueryBinder starting");

        self.bind = function(){
            RestApiService.get("query/get?code="+self.queryCode).
            success(function(data) {
                self.values = data.items;
            })
        };;

        self.$onInit = function() {
            if (self.bindOnInit) {
                self.bind();
            }
            console.log("subscribing to bindBuComponents");
            PubSub.subscribe("bindBuComponents",self.bind);
        };

    }
});

MetronicApp.component('duSelectButtons', {
    template: "<div class=\"form-group\"><label>{{$ctrl.label}}</label>"
    +
    "<div><button type=\"button\" ng-click=\"$ctrl.setValue(value)\" ng-class=\"$ctrl.idModel == value.id?'btn btn-primary':'btn'\" ng-repeat=\"value in $ctrl.values\">{{value.name}}</button></div></div>" +
    "",
    bindings: {
        queryCode: '@',
        label: '@',
        ngModel: '<',
        idModel : '=',

    },
    require: { ngModelCtrl: 'ngModel' },
    controller: function (RestApiService,PubSub) {
        var self = this;

        console.log("buQueryBinder starting");
        self.setValue = function(value){
            //self.ngModel = ngModel;
            console.log("setValue",value);
            self.idModel = value.id;
            self.ngModel = value;
            self.ngModelCtrl.$setViewValue(self.ngModel);
        }

        self.ngModelChange = function () {
            self.idModel = self.ngModel.id;
            self.ngModelCtrl.$setViewValue(this.ngModel);

        };
        self.bind = function(){
            RestApiService.get("query/get?code="+self.queryCode).
            success(function(data) {
                self.values = data.items;
            })
        };;

        self.$onInit = function() {
                self.bind();

        };

    }
});

