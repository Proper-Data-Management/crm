MetronicApp.service('dealService', function() {
    var accountId = 0;
    var Id = 0;

    var setAccountId = function(i) {
        accountId = i
    };

    var getAccountId = function(){
        return accountId;
    };

    var setId = function(i) {
        Id = i
    };

    var getId = function(){
        return Id;
    };


    return {
        setAccountId: setAccountId,
        getAccountId: getAccountId,
        setId: setId,
        getId: getId
    };

});