
/* Setup general page controller */
MetronicApp.controller('ModalDetailPageController', ['$scope', '$stateParams','PubSub',



    function ($scope, $stateParams,PubSub) {



        //$scope.modals = [{id:'lookupModal1'},{id:'lookupModal2'},{id:'lookupModal3'}];
        $scope.modals = [];
        for (i=0;i<=100;i++){

            $scope.modals.push({padding:i*50,id:'lookupModal'+i});
        }
        $scope.current = 0;
        $scope.stoppedRefresh = false;
        $scope.detail = {};

        $scope.choose = function(id){
            $scope.sender.idModel = id;
            if ( typeof($scope.sender.onChange)!="undefined" ) {
                $scope.sender.onChange($scope.sender);
            }
            //$scope.close();
        }

        $scope.close = function(i){
            $("#"+i.id).modal('hide');
            $scope.current --;
        }

        $scope.location = function(url){
            location.href=url;
        }


        $scope.changeHash = function(){

            var off = $scope.$on('$stateChangeStart', function(e) {
                e.preventDefault();
            });
            off();
            //$location.path('product/123').replace();
        }

        $scope.open = function (init){

            $scope.current ++;

            newId = 'lookupModal'+($scope.current);


            //window.location.hash = "dsda";
            //$scope.changeHash();

            /*if (init.hash){

             history.pushState(null, null, init.hash);

             }*/

            $scope.modals[$scope.current].url = "/restapi/pagetemplate?code="+init.sender.lookupPageCode+"&pk="+init.data.id;
            $scope.modals[$scope.current].title = init.title?init.title:init.data.title;
            $stateParams.id = init.data.id;

            $('#'+newId).modal({
                backdrop: 'static',
                keyboard: false
            });
            $scope.sender = init.sender;


        }

        var sub = PubSub.subscribe('openDetailModal', $scope.open);
        //alert("test11");




}]);

