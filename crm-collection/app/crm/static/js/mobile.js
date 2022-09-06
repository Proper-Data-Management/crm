/***
Metronic AngularJS App Main Script
***/

var $stateProviderRef = null;

/* Metronic App */
var MetronicApp = angular.module("MetronicApp", [
    "ui.router",
    "ui.bootstrap",
    "oc.lazyLoad",
    "ngSanitize",
    "pascalprecht.translate",
    'smart-table',
    'angularFileUpload',
    "ui.mask",
    "tmh.dynamicLocale",
    "ui.checkbox",
    "flowChart",
    "angular.css.injector",
    "ui.ace",
    "ui.select",
    "ngMap",
    "ngLocale",
    "ui.tree",
    "PubSub",
    "colorpicker.module",
    "easypiechart",
    "ckeditor",
    "treeControl",
    "chart.js",
    "angularMoment",
    "ng-webcam",
	"ngTouch"
])

.filter('startFrom', function () {
	return function (input, start) {
		if (input) {
			start = +start;
			return input.slice(start);
		}
		return [];
	};
})

.filter('propsFilter', function() {
    return function(items, props) {
        var out = [];

        if (angular.isArray(items)) {
            items.forEach(function(item) {
                var itemMatches = false;

                var keys = Object.keys(props);
                for (var i = 0; i < keys.length; i++) {
                    var prop = keys[i];
                    var text = props[prop].toLowerCase();
                    if (item[prop].toString().toLowerCase().indexOf(text) !== -1) {
                        itemMatches = true;
                        break;
                    }
                }

                if (itemMatches) {
                    out.push(item);
                }
            });
        } else {
            // Let the output be the input untouched
            out = items;
        }

        return out;
    };
})
    .filter('split', function() {
        return function(input, splitChar, splitIndex) {
            // do some bounds checking here to ensure it has that index
            return input.split(splitChar)[splitIndex];
        }
    })
    .filter('phone', function () {
    return function (tel) {
        if (!tel) { return ''; }

        var value = tel.toString().trim().replace(/^\+/, '');

        if (value.match(/[^0-9]/)) {
            return tel;
        }

        var country, city, number;

        switch (value.length) {
            case 10: // +1PPP####### -> C (PPP) ###-####
                country = 1;
                city = value.slice(0, 3);
                number = value.slice(3);
                break;

            case 11: // +CPPP####### -> CCC (PP) ###-####
                country = value[0];
                city = value.slice(1, 4);
                number = value.slice(4);
                break;

            case 12: // +CCCPP####### -> CCC (PP) ###-####
                country = value.slice(0, 3);
                city = value.slice(3, 5);
                number = value.slice(5);
                break;

            default:
                return tel;
        }

        if (country == 1) {
            country = "";
        }

        number = number.slice(0, 3) + '-' + number.slice(3);

        return (country + " (" + city + ") " + number).trim();
    };
})
    .filter('range', function () {
        return function(n) {
            var res = [];
            for (var i = 0; i < n; i++) {
                res.push(i);
            }
            return res;
        };
    })

    .filter('tel', function () {
        return function (tel) {

            if (tel) {
                if (tel.length == 11) {
                    return "8" + tel.substring(1);
                }
                else {
                    return tel;
                }
            }
    }})
    .filter('num2word',function(){
        String.prototype.trimRight=function(){
            // убирает все пробелы в конце строки
            var r=/\s+$/g;
            return this.replace(r,'');
        }
        String.prototype.trimLeft=function(){
            // убирает все пробелы в начале строки
            var r=/^\s+/g;
            return this.replace(r,'');
        }
        String.prototype.trim=function(){
            // убирает все пробелы в начале и в конце строки
            return this.trimRight().trimLeft();
        }
        String.prototype.trimMiddle=function(){
            // убирает все пробелы в начале и в конце строки
            // помимо этого заменяет несколько подряд
            // идущих пробелов внутри строки на один пробел
            var r=/\s\s+/g;
            return this.trim().replace(r,' ');
        }
        String.prototype.trimAll=function(){
            // убирает все пробелы в строке s
            var r=/\s+/g;
            return this.replace(r,'');
        }
        String.prototype.repeat=function(n){
            // повторяет строку n раз
            var r='';
            if(typeof(n)=='number')
                for(var i=1; i<=n; i++) r+=this;
            return r;
        }
        Number.prototype.isInt=function(){
            // возвращает True, если число является целым
            return (Math.round(this)==this);
        }
        Number.prototype.roundTo=function(n){
            // округляет число до заданного количества n
            // знаков после (или перед) запятой
            var x=0;
            if(typeof(n)=='number')
                if(n.isInt())
                    if(n>= -6 && n<=6) x=n;
            x=Math.pow(10,x);
            return Math.round(this*x)/x;
        }

        String.prototype.trimAll=function(){
            // убирает все пробелы в строке s
            var r=/\s+/g;
            return this.replace(r,'');
        }

        Number.prototype.toPhrase=function(c){
            // сумма прописью для чисел от 0 до 999 триллионов
            // можно передать параметр "валюта": RUB,USD,EUR (по умолчанию RUB)
            var x=this.roundTo(2);
            if(x<0 || x>999999999999999.99) return false;
            var currency='KZT';
            if(typeof(c)=='string') currency=c.trimAll().toUpperCase();
            if(currency=='RUR') currency='RUB';
            if(currency!='RUB' && currency!='USD' && currency!='EUR'  && currency!='KZT') return false;
            var groups=new Array();
            groups[0]=new Array();
            groups[1]=new Array();
            groups[2]=new Array();
            groups[3]=new Array();
            groups[4]=new Array();
            groups[9]=new Array();
            // рубли
            // по умолчанию
            groups[0][-1]={'KZT':'тенге','RUB':'рублей','USD':'долларов США','EUR':'евро'};
            //исключения
            groups[0][1]={'KZT':'тенге','RUB':'рубль','USD':'доллар США','EUR':'евро'};
            groups[0][2]={'KZT':'тенге','RUB':'рубля','USD':'доллара США','EUR':'евро'};
            groups[0][3]={'KZT':'тенге','RUB':'рубля','USD':'доллара США','EUR':'евро'};
            groups[0][4]={'KZT':'тенге','RUB':'рубля','USD':'доллара США','EUR':'евро'};
            // тысячи
            // по умолчанию
            groups[1][-1]='тысяч';
            //исключения
            groups[1][1]='тысяча';
            groups[1][2]='тысячи';
            groups[1][3]='тысячи';
            groups[1][4]='тысячи';
            // миллионы
            // по умолчанию
            groups[2][-1]='миллионов';
            //исключения
            groups[2][1]='миллион';
            groups[2][2]='миллиона';
            groups[2][3]='миллиона';
            groups[2][4]='миллиона';
            // миллиарды
            // по умолчанию
            groups[3][-1]='миллиардов';
            //исключения
            groups[3][1]='миллиард';
            groups[3][2]='миллиарда';
            groups[3][3]='миллиарда';
            groups[3][4]='миллиарда';
            // триллионы
            // по умолчанию
            groups[4][-1]='триллионов';
            //исключения
            groups[4][1]='триллион';
            groups[4][2]='триллиона';
            groups[4][3]='триллиона';
            groups[4][4]='триллиона';
            // копейки
            // по умолчанию
            groups[9][-1]={'KZT':'тиын','RUB':'копеек','USD':'центов','EUR':'центов'};
            //исключения
            groups[9][1]={'KZT':'тиын','RUB':'копейка','USD':'цент','EUR':'цент'};
            groups[9][2]={'KZT':'тиын','RUB':'копейки','USD':'цента','EUR':'цента'};
            groups[9][3]={'KZT':'тиын','RUB':'копейки','USD':'цента','EUR':'цента'};
            groups[9][4]={'KZT':'тиын','RUB':'копейки','USD':'цента','EUR':'цента'};
            // цифры и числа
            // либо просто строка, либо 4 строки в хэше
            var names=new Array();
            names[1]={0:'один',1:'одна',2:'один',3:'один',4:'один'};
            names[2]={0:'два',1:'две',2:'два',3:'два',4:'два'};
            names[3]='три';
            names[4]='четыре';
            names[5]='пять';
            names[6]='шесть';
            names[7]='семь';
            names[8]='восемь';
            names[9]='девять';
            names[10]='десять';
            names[11]='одиннадцать';
            names[12]='двенадцать';
            names[13]='тринадцать';
            names[14]='четырнадцать';
            names[15]='пятнадцать';
            names[16]='шестнадцать';
            names[17]='семнадцать';
            names[18]='восемнадцать';
            names[19]='девятнадцать';
            names[20]='двадцать';
            names[30]='тридцать';
            names[40]='сорок';
            names[50]='пятьдесят';
            names[60]='шестьдесят';
            names[70]='семьдесят';
            names[80]='восемьдесят';
            names[90]='девяносто';
            names[100]='сто';
            names[200]='двести';
            names[300]='триста';
            names[400]='четыреста';
            names[500]='пятьсот';
            names[600]='шестьсот';
            names[700]='семьсот';
            names[800]='восемьсот';
            names[900]='девятьсот';
            var r='',i,j,y=Math.floor(x);
            // если НЕ ноль рублей
            if(y>0){
                // выделим тройки с руб., тыс., миллионами, миллиардами и триллионами
                var t=new Array();
                for(i=0; i<=4; i++){
                    t[i]=y%1000;
                    y=Math.floor(y/1000);
                }
                var d=new Array();
                // выделим в каждой тройке сотни, десятки и единицы
                for(i=0; i<=4; i++){
                    d[i]=new Array();
                    d[i][0]=t[i]%10; // единицы
                    d[i][10]=t[i]%100-d[i][0]; // десятки
                    d[i][100]=t[i]-d[i][10]-d[i][0]; // сотни
                    d[i][11]=t[i]%100; // две правых цифры в виде числа
                }
                for(i=4; i>=0; i--)
                    if(t[i]>0){
                        if(names[d[i][100]]) r+=' '+((typeof(names[d[i][100]])=='object')?(names[d[i][100]][i]):(names[d[i][100]]));
                        if(names[d[i][11]]) r+=' '+((typeof(names[d[i][11]])=='object')?(names[d[i][11]][i]):(names[d[i][11]]));
                        else{
                            if(names[d[i][10]]) r+=' '+((typeof(names[d[i][10]])=='object')?(names[d[i][10]][i]):(names[d[i][10]]));
                            if(names[d[i][0]]) r+=' '+((typeof(names[d[i][0]])=='object')?(names[d[i][0]][i]):(names[d[i][0]]));
                        }
                        // если существует числительное
                        if(names[d[i][11]]) j=d[i][11];
                        else j=d[i][0];
                        if(groups[i][j])
                            if(i==0) r+=' '+groups[i][j][currency];
                            else r+=' '+groups[i][j];
                        else{
                            if(i==0) r+=' '+groups[i][-1][currency];
                            else r+=' '+groups[i][-1];
                        }
                    }
                if(t[0]==0) r+=' '+groups[0][-1][currency];
            }else r='Ноль '+groups[0][-1][currency];
            y=((x-Math.floor(x))*100).roundTo();
            if(y<10) y='0'+y;
            r=r.trimMiddle();
            r=r.substr(0,1).toUpperCase()+r.substr(1);
            r+=' '+y;
            y=y*1;
            // если существует числительное
            if(names[y]) j=y;else j=y%10;
            if(groups[9][j]) r+=' '+groups[9][j][currency];
            else r+=' '+groups[9][-1][currency];
            return r;
        }
        return function(input){ return (new Number(input)).toPhrase('KZT'); }
    });

/* Configure ocLazyLoader(refer: https://github.com/ocombe/ocLazyLoad) */
MetronicApp.config(['$ocLazyLoadProvider', function($ocLazyLoadProvider) {
    $ocLazyLoadProvider.config({
        // global configs go here
		 
    });
}]);

MetronicApp.config(function($sceDelegateProvider) {
    $sceDelegateProvider.resourceUrlWhitelist([
      'self',
      'https://www.zvanda.kz/**'
    ]);
  });

  


/********************************************
 BEGIN: BREAKING CHANGE in AngularJS v1.3.x:
*********************************************/
/**
`$controller` will no longer look for controllers on `window`.
The old behavior of looking on `window` for controllers was originally intended
for use in examples, demos, and toy apps. We found that allowing global controller
functions encouraged poor practices, so we resolved to disable this behavior by
default.

To migrate, register your controllers with modules rather than exposing them
as globals:

Before:

```javascript
function MyController() {
  // ...
}
```

After:

```javascript
angular.module('myApp', []).controller('MyController', [function() {
  // ...
}]);

Although it's not recommended, you can re-enable the old behavior like this:

```javascript
angular.module('myModule').config(['$controllerProvider', function($controllerProvider) {
  // this option might be handy for migrating old apps, but please don't use it
  // in new ones!
  $controllerProvider.allowGlobals();
}]);
**/



//AngularJS v1.3.x workaround for old style controller declarition in HTML
MetronicApp.config(['$controllerProvider', function($controllerProvider) {
  // this option might be handy for migrating old apps, but please don't use it
  // in new ones!
  $controllerProvider.allowGlobals();
}]);

/********************************************
 END: BREAKING CHANGE in AngularJS v1.3.x:
*********************************************/

/* Setup global settings */
MetronicApp.factory('settings', ['$rootScope', function($rootScope) {
    // supported languages
    var settings = {
        layout: {
            pageSidebarClosed: false, // sidebar menu state
            pageBodySolid: false, // solid body color state
            pageAutoScrollOnLoad: 1000 // auto scroll to top on page load
        },
        layoutImgPath: Metronic.getAssetsPath() + 'admin/layout/img/',
        layoutCssPath: Metronic.getAssetsPath() + 'admin/layout/css/'
    };

    $rootScope.settings = settings;
    $rootScope.mainuri = "/static";

    return settings;
}]);

/* Setup App Main Controller */
MetronicApp.controller('AppController', ['$scope', '$rootScope', function($scope, $rootScope) {
    $scope.$on('$viewContentLoaded', function() {
        Metronic.initComponents(); // init core components
        //Layout.init(); //  Init entire layout(header, footer, sidebar, etc) on page load if the partials included in server side instead of loading with ng-include directive
    });
    //$scope.changeLanguage("ru");
}]);

/***
Layout Partials.
By default the partials are loaded through AngularJS ng-include directive. In case they loaded in server side(e.g: PHP include function) then below partial
initialization can be disabled and Layout.init() should be called on page load complete as explained above.
***/

/* Setup Layout Part - Header */
MetronicApp.controller('HeaderController', ['$scope', function($scope) {
    $scope.$on('$includeContentLoaded', function() {
        Layout.initHeader(); // init header
    });
}]);

/* Setup Layout Part - Sidebar */
MetronicApp.controller('SidebarController', ['$scope', function($scope) {
    $scope.$on('$includeContentLoaded', function() {
        Layout.initSidebar(); // init sidebar
    });
}]);

/* Setup Layout Part - Quick Sidebar */
MetronicApp.controller('QuickSidebarController', ['$scope', function($scope) {
    $scope.$on('$includeContentLoaded', function() {
        setTimeout(function(){
            QuickSidebar.init(); // init quick sidebar
        }, 2000)
    });
}]);

/* Setup Layout Part - Theme Panel */
MetronicApp.controller('ThemePanelController', ['$scope', function($scope) {
    $scope.$on('$includeContentLoaded', function() {
        Demo.init(); // init theme panel
    });
}]);

/* Setup Layout Part - Footer */
MetronicApp.controller('FooterController', ['$scope', function($scope) {



    $scope.$on('$includeContentLoaded', function() {
        Layout.initFooter(); // init footer
    });
}]);

/* Setup Rounting For All Pages */
MetronicApp.config(['$stateProvider', '$urlRouterProvider', function($stateProvider, $urlRouterProvider) {
    // Redirect any unmatched url


    /*$urlRouterProvider.rule(function ($injector, $location) {
        //what this function returns will be set as the $location.url
        var path = $location.path(), normalized = path.toLowerCase();
        if (path != normalized) {
            //instead of returning a new url string, I'll just change the $location.path directly so I don't have to worry about constructing a new url string and so a new state change is not triggered
            $location.replace().path(normalized);

        }
        //$location.replace().path("#123");
        alert($location.path());

        // because we've returned nothing, no state change occurs
    });*/

    $urlRouterProvider.deferIntercept();




    $urlRouterProvider.otherwise("/404");
    $stateProviderRef = $stateProvider;
}]);




/*
    $.getJSON( "js/i18n/translate.json", function( data ) {
        async: false,

        //console.log(data);

        MetronicApp.config(function($translateProvider) {
            $translateProvider.translations('en', data["en"]).translations('ru', data["ru"]);
            $translateProvider.preferredLanguage('en');

        });







});
 */







MetronicApp.run(['$rootScope', '$urlRouter', '$state','$translate', '$http', '$log','tmhDynamicLocale','amMoment',
    function ($rootScope, $urlRouter, $state,$translate, $http, $log,tmhDynamicLocale,amMoment) {





    amMoment.changeLocale('ru');

    $rootScope.changeLanguage = function (langKey) {
        $translate.use(langKey);
        tmhDynamicLocale.set(langKey);
        amMoment.changeLocale(langKey);

	$http.post('/auth/setLanguage',{lang: langKey}).
	success(function(data) {

	}).error(function(data, status){
        console.log("error on setLanguage", status, data);
    });;

    };


    $rootScope.$on('$locationChangeStart', function( event ) {
        if ($rootScope.promptExit) {
            var answer = confirm('If you leave this page you are going to lose all unsaved changes, are you sure you want to leave?')
            if (!answer) {
                event.preventDefault();
            }else{
                $rootScope.promptExit = false;
            }
        }
    });


}]);




///* Init global settings and run the app */
MetronicApp.run(['$urlRouter',"$rootScope", '$http',"settings", "$state","cssInjector", "$timeout", "RestApiService", "$location",

    function($urlRouter,$rootScope, $http, settings, $state,cssInjector,$timeout,RestApiService,$location) {

    //Open ACcount
    $rootScope.processData = function(data){

        if (data.sessioninfo == null) {
            location.href = "/auth/logout?angularjs_redirecturi="+ $location.url();
            //alert('Exit');
        }
        localStorage.setItem("sessioninfo",JSON.stringify(data));
        $rootScope.sessioninfo = data.sessioninfo;
        $rootScope.isMobile = true;
        $rootScope.lang = "ru";
        //$rootScope.uri = data.uri;
        
        if (data.session_parameters){
            $rootScope.session_parameters = [];
            angular.forEach(data.session_parameters, function (value, key) {
                $rootScope.session_parameters[value.code] = value.value;
            });

            if ($rootScope.session_parameters.custom_css) {
                    cssInjector.add($rootScope.session_parameters.custom_css);
            }
        
        }


        $rootScope.session_roles = {};
        if (data.session_roles) {				
            angular.forEach(data.session_roles, function (value, key) {
                $rootScope.session_roles[value.code] = true;
                //console.log(value.code);
            });	
        }				
        $rootScope.sessionRoleParams = [];
        if (data.session_role_params) {
            angular.forEach(data.session_role_params, function (value, key) {
                $rootScope.sessionRoleParams[value.code] = value.value;
            });			
        }				
        

RestApiService
    .get("pages/get")
    .success(function(data) {
        angular.forEach(data, function(value, key) {


            //console.log(value);
            var getExistingState = $state.get(value.name)

            if(getExistingState !== null){
                return;
            }




            //value["files"] = ["theme/assets/global/plugins/select2/select2.css","theme/assets/global/plugins/datatables/plugins/bootstrap/dataTables.bootstrap.css","theme/assets/global/plugins/bootstrap-datepicker/js/bootstrap-datepicker.min.js","theme/assets/global/plugins/select2/select2.min.js","theme/assets/global/plugins/datatables/all.min.js","theme/assets/global/scripts/datatable.js","js/scripts/table-ajax.js","js/controllers/GeneralPageController.js","js/plugins/ui-select/select.min.js","js/plugins/ui-select/select.min.css","js/plugins/ckeditor/ckeditor.js"];
            
            value["files"] = ["js/controllers/GeneralPageController.js","js/plugins/ui-select/select.min.js","js/plugins/ui-select/select.min.css"];
            
            var files = [];
            //alert("test");
            angular.forEach(value["files"], function(value, key) {
                //console.log(value);
                files.push(value);
            });

        data={
            url: value.url,
            templateUrl: value.templateurl,
            data: {params: value.params, pageTitle: value.title,pageCode: value.code,entityCode: value.entity_code,entityId: value.entity_id,pageId: value.id,pageIcon: value.icon,pageQueryCode: value.query_code,pageQueryId: value.query_id,pageFilterSetCode: value.filter_set_code,pageFilterSetId:value.filter_set_id},
            controller: value.controller,
            resolve: {
                deps: ['$ocLazyLoad', function($ocLazyLoad) {
                    return $ocLazyLoad.load({
                        name: 'MetronicApp',
                        insertBefore: '#ng_load_plugins_before', // load the above css files before a LINK element with this ID. Dynamic CSS files must be loaded between core and theme css files
                        files: files
                    });
                }]
            }
        };

        dataPage={
            url: value.url+"/?p[]",
            templateUrl: value.templateurl,
            data: {params: value.params, pageTitle: value.title,pageCode: value.code,entityCode: value.entity_code,entityId: value.entity_id,pageId: value.id,pageIcon: value.icon,pageQueryCode: value.query_code,pageQueryId: value.query_id,pageFilterSetCode: value.filter_set_code,pageFilterSetId:value.filter_set_id},
            controller: value.controller,
            resolve: {
                deps: ['$ocLazyLoad', function($ocLazyLoad) {
                    return $ocLazyLoad.load({
                        name: 'MetronicApp',
                        insertBefore: '#ng_load_plugins_before', // load the above css files before a LINK element with this ID. Dynamic CSS files must be loaded between core and theme css files
                        files: files
                    });
                }]
            }
        };


            $stateProviderRef

            // Dashboard
                .state(value.code, data)
                .state("syspage2_"+value.code, dataPage)

        });
        // Configures $urlRouter's listener *after* your custom listener





        $urlRouter.sync();
        $urlRouter.listen();






    });				
        
        $.ajax({
            url: "../restapi/translates/get",
            dataType: 'json',
            async: false,

            success: function(data) {
                //stuff
                //...
                MetronicApp.config(function($translateProvider) {
                    $translateProvider.useSanitizeValueStrategy(null);
                    $translateProvider.translations('en', data["en"]).translations('ru', data["ru"]).translations('kk', data["kk"]);
                    $translateProvider.preferredLanguage(data.lang);
                    //alert($rootScope.lang);
                    $rootScope.changeLanguage(data.lang);

                });
            }
        });        
    } ; 


    $rootScope.navigator = navigator;




    $rootScope.createAccount= function(caller_id){
        $rootScope.waitCall.show = false;
        window.location.href="#/src/accountdetails/0?set_caller_id="+caller_id;
        location.reload();
    }

    $rootScope.openAccount= function(accountId){
        $rootScope.waitCall.show = false;
        window.location.href="#/src/accountdetails/"+accountId;
        location.reload();
    }

        $rootScope.$state = $state; // state to be accessed from view
        //$rootScope.$state = $state; // state to be accessed from view
        //var $state = $rootScope.$state;


    if (location.pathname != "/static/p.htm") {
		

        RestApiService
            .get("services/run/session")

            .error(function(data, status){
                console.log("error on sessioninfo999", status, data);
                console.log("sessioninfo",localStorage.getItem("sessioninfo"));




////////////////////////// OFFLINE

if (localStorage.getItem("sessioninfo")!=null){
var data  = JSON.parse(localStorage.getItem("sessioninfo"));
$rootScope.processData(data);
}
////////////////////////////////OFFLINE
            })            
            .success(function (data) {
                //alert(data.items);
            $rootScope.processData(data);        

            }).error(function(data, status){
                console.log("error on sessioninfo", status, data);
            });		
		
		
		
		
        /*RestApiService
            .get("query/get?code=sessioninfo")
            .success(function (data) {
                //alert(data.items);
                if (data.items == null) {
                    location.href = "/auth/logout?angularjs_redirecturi="+ $location.url();
                    //alert('Exit');
                }
                $rootScope.sessioninfo = data.items[0];
                $rootScope.isMobile = data.isMobile;
                $rootScope.lang = data.lang;
                $rootScope.uri = data.uri;



            }).error(function(data, status){
                console.log("error on sessioninfo", status, data);
            });*/




        /*RestApiService
            .get("query/get?code=session_parameters")
            .success(function (data) {

                $rootScope.session_parameters = [];
                angular.forEach(data.items, function (value, key) {
                    $rootScope.session_parameters[value.code] = value.value;
                });

                if ($rootScope.session_parameters.custom_css) {
                    cssInjector.add($rootScope.session_parameters.custom_css);
                }

            }).error(function(data, status){
                console.log("error on session_parameters", status, data);
            });

        RestApiService
            .get("query/get?code=session_role_params")
            .success(function (data) {

                $rootScope.sessionRoleParams = [];
                angular.forEach(data.items, function (value, key) {
                    $rootScope.sessionRoleParams[value.code] = value.value;
                });
            }).error(function(data, status){
                console.log("error on session_role_params", status, data);
            });

        RestApiService
            .get("query/get?code=session_roles")
            .success(function (data) {

                $rootScope.session_roles = [];
                angular.forEach(data.items, function (value, key) {
                    $rootScope.session_roles[value.code] = true;
                    //console.log(value.code);
                });
                


            }).error(function(data, status) {
                console.error('error on session_roles', status, data);
              });*/
    }


    
    }
]);



MetronicApp.config( [
    '$compileProvider',
    function( $compileProvider )
    {
        $compileProvider.aHrefSanitizationWhitelist(/^\s*(https?|sip|mailto|chrome-extension|data|tel):/);
        // Angular before v1.2 uses $compileProvider.urlSanitizationWhitelist(...)
    }
]);