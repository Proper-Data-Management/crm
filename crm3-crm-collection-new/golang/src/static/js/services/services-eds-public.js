MetronicApp.service('NCALayer', ['$q', '$rootScope', function($q, $rootScope) {
    // We return this object to anything injecting our service

    var NCALayer = {};

    var heartbeatMsg = '--heartbeat--';
    var heartbeatInterval = null;
    var missedHeartbeats = 0;
    var missedHeartbeatsLimitMin = 3;
    var missedHeartbeatsLimitMax = 50;
    var missedHeartbeatsLimit = missedHeartbeatsLimitMin;

    // Create our websocket object with the address to the websocket
    var ws = null;
    var answer = {};
    var edsKeys = [];

    var storageAlias = 'PKCS12';
    var ncaPassword = "";
    var storagePath = "";
    var person = [];

    var data = {
        'method': "",
        'args': []
    };

    var getEDSData = function() {
        missedHeartbeats = missedHeartbeatsLimitMax;
        // Storing in a variable for clarity on what sendRequest returns
        var promise = sendRequest(data);

        return promise;
    };

    /** selectSignType() определеяет тип подписки Java апплет или прослойка
     *====================================================================*/
    var selectSignType = function () {
        if($rootScope.isMobile) {
            var result = Mobile.fileExists()
            if(typeof result != "undefined") {
                if(result) {
                    NCALayer.showPassword(true);
                }
            }
        } else {
            data.method = "browseKeyStore";
            data.args = [storageAlias, 'P12', storagePath];
            getEDSData();
            ncaPassword = "";
        }
    };

    /** setNCAPassword() - вызываем после ввода пароля и отправки
     *====================================================================*/
    var setNCAPassword = function (password) {
        if (password) {
            ncaPassword = password;
            if($rootScope.isMobile) {
                var result = Mobile.getSubjectDN(ncaPassword);
                if(typeof result != "undefined") {
                    if(result) {
                        NCALayer.showPerson(true);
                        NCALayer.postGetSubjectDN({result : result})
                    } else {
                        NCALayer.showError({ errorCode : "WRONG_PASSWORD" });
                    }
                }
            } else {
                data.method = "getKeys";
                data.args = [storageAlias, storagePath, ncaPassword, 'SIGN'];
                getEDSData();
            }
        } else {
            alert('Введите пароль');
        }
    };

    /** signXml() - подписываем XML
     *====================================================================*/
    var signXml = function (xmlData) {

        //var xmlData = json2xml(xmlData,'root');
        if($rootScope.isMobile) {
            NCALayer.showPassword(false);
            NCALayer.showPerson(false);
            var result = Mobile.doSignature(xmlData,ncaPassword);
            if(typeof result != "undefined") {
                NCALayer.postSignXml({ result : result });
            }
        } else {
            data.method = "signXml";
            data.args = [
                storageAlias,
                storagePath,
                edsKeys[0][3],
                ncaPassword,
                xmlData
            ];
            console.log("ОТПРАВЛЕНО!")
            getEDSData();
        }
    };

    /** showFileChooser() - выбрать подписываем файл
     *====================================================================*/
    var showFileChooser = function (fileToDownload) {

        if($rootScope.isMobile) {
            console.log("Запушено окно выбора файла для андроид!");
            NCALayer.showPassword(false);
            NCALayer.showPerson(false);
            var result = Mobile.doDownloadFile(fileToDownload);
            if(typeof result != "undefined") {
                NCALayer.postShowFileChooser({ result : result });
            }
        } else {
            data.method = "showFileChooser";
            data.args = ["ALL",""];
            console.log("Запушено окно выбора файла!");
            getEDSData();
        }
    };

    /** createCMSSignatureFromFile() - подписать файл
     *====================================================================*/
    var createCMSSignatureFromFile = function (fileSignPath) {

        if($rootScope.isMobile) {
            NCALayer.showPassword(false);
            NCALayer.showPerson(false);
            var result = Mobile.doSignatureFile(fileSignPath,ncaPassword);
            if(typeof result != "undefined") {
                NCALayer.postCreateCMSSignatureFromFile({ result : result });
            }
        } else {
            data.method = "createCMSSignatureFromFile";
            data.args = [
                storageAlias,
                storagePath,
                edsKeys[0][3],
                ncaPassword,
                fileSignPath,
                false // Присоединение данных файла
            ];
            console.log("ОТПРАВЛЕНО!")
            getEDSData();
        }
    };


    function init(request) {
        ws = new WebSocket("wss://127.0.0.1:13579/");

        ws.onopen = function () {
            console.log("Socket has been opened!");
            if (heartbeatInterval === null) {
                missedHeartbeats = 0;
                heartbeatInterval = setInterval(pingNCALayer, 1000);
            }

            if(request != null) {
                sendRequest(request);
            }
        };

        ws.onmessage = function (response) {
            listener(response.data)
        };

        ws.onclose = function (event) {
            if (!event.wasClean) {
                NCALayer.wsError();
                console.log('Ошибка при подключений к прослойке');
            } else {
                NCALayer.wsClosed();
                console.log('Отключено!');
            }
        };
    }

    function listener(str) {

        console.log("Ответ: " + str);

        if(str == heartbeatMsg) {
            return;
        }

        answer = JSON.parse(str);

        if(answer.errorCode !== undefined) {
            if (answer.errorCode === 'NONE')
            {
                switch (data.method) {
                    case 'browseKeyStore':
                        storagePath = answer.result;
                        console.log(data.method + " " + answer.errorCode);
                        if (storagePath) {
                            NCALayer.showPassword(true);
                            missedHeartbeatsLimit = missedHeartbeatsLimitMin;
                        } else {
                            console.log("Не выбран путь к файлу");
                        }
                        break;
                    case 'getKeys':
                        postGetKey();
                        break;
                    case 'getSubjectDN' :
                        NCALayer.postGetSubjectDN(answer);
                        data.method = 'getNotBefore';
                        data.args = [storageAlias, storagePath, edsKeys[0][3], ncaPassword];
                        getEDSData();
                        break;
                    case 'getNotBefore':
                        NCALayer.postGetDate(answer,'beginDate');
                        data.method = 'getNotAfter';
                        data.args = [storageAlias, storagePath, edsKeys[0][3], ncaPassword];
                        getEDSData();
                        break;
                    case 'getNotAfter':
                        NCALayer.postGetDate(answer,'endDate');
                        break;
                    case 'signXml' :
                        NCALayer.postSignXml(answer);
                        NCALayer.showPerson(false);
                        break;
                    case 'showFileChooser' :
                        NCALayer.postShowFileChooser(answer);
                        break;
                    case 'createCMSSignatureFromFile' :
                        NCALayer.postCreateCMSSignatureFromFile(answer);
                        NCALayer.showPerson(false);
                        break;
                }
            } else {
                NCALayer.showError(answer);
            }
        }
    }

    function postGetKey() {

        var slotListArr = answer.result.split('\n');

        if (slotListArr.length > 0) {
            for (var counter = 0; counter < slotListArr.length; counter++) {
                if (slotListArr[counter] === null || slotListArr[counter] === '') {
                    continue;
                }
                edsKeys.push(slotListArr[counter].split('|'));
            }

            data.method = 'getSubjectDN';
            data.args = [
                storageAlias,
                storagePath,
                edsKeys[0][3],
                ncaPassword
            ];
            console.log("storageAlias = " + storageAlias);
            console.log("storagePath = " + storagePath);
            console.log("edsKeys = " + edsKeys[0][3]);
            getEDSData();
        } else {
            // $scope.ncaPassword_message_class = '';
        }

        NCALayer.showPassword(false);

    }


    function pingNCALayer() {
        try {
            // missedHeartbeats++;
            //
            // if (missedHeartbeats >= missedHeartbeatsLimit) {
            //    throw new Error('Too many missed heartbeats.');
            // }
            ws.send(heartbeatMsg);
        } catch (error) {
            clearInterval(heartbeatInterval);
            heartbeatInterval = null;
            ws.close();
        }
    }

    function sendRequest(request) {
        /**
         * CONNECTING         0     The connection is not yet open.
         * OPEN               1     The connection is open and ready to communicate.
         * CLOSING            2     The connection is in the process of closing.
         * CLOSED             3     The connection is closed or couldn't be opened.
         */
        if (ws === null || ws.readyState === 3 || ws.readyState === 2) {
            return init(request);
        } else {
            var defer = $q.defer();
            console.log('Sending request', request);
            ws.send(JSON.stringify(request));
        }

        return defer.promise;
    }

    var json2xml = (function () {

        "use strict";

        var tag = function (name, closing) {
            return "<" + (closing ? "/" : "") + name + ">";
        };

        return function (obj, rootname) {
            var xml = "";
            for (var i in obj) {
                if (obj.hasOwnProperty(i)) {
                    var value = obj[i],
                        type = typeof value;
                    if (value instanceof Array && type == 'object') {
                        for (var sub in value) {
                            xml += json2xml(value[sub]);
                        }
                    } else if (value instanceof Object && type == 'object') {
                        xml += tag(i) + json2xml(value) + tag(i, 1);
                    } else {
                        xml += tag(i) + value + tag(i, {
                                closing: 1
                            });
                    }
                }
            }

            return rootname ? tag(rootname) + xml + tag(rootname, 1) : xml;
        };
    })(json2xml || {});

    NCALayer.showError = function(answer) {
        if (answer.errorCode === 'WRONG_PASSWORD' && answer.result > -1) {
            alert('Пароль неверен. Осталось попыток: ' + answer.result);
        } else if (answer.errorCode === 'WRONG_PASSWORD') {
            alert('Пароль неверен');
        } else {
            if (answer.errorCode === 'EMPTY_KEY_LIST') {
                alert('В хранилище не найдено подходящих сертификатов для ЭЦП')
            } else {
                alert('Код ошибки: ' + answer.errorCode);
            }
        }
    };

    NCALayer.wsClosed = function() {
        //clear
    };

    NCALayer.wsError = function() {
        //
    };

    NCALayer.bind = function($theScope)
    {
        $theScope.person = person;
        $theScope.selectSignType = selectSignType;
        $theScope.setNCAPassword = setNCAPassword;
        $theScope.signXml = signXml;
        $theScope.showFileChooser = showFileChooser;
        $theScope.createCMSSignatureFromFile = createCMSSignatureFromFile;
    };

    return NCALayer;
}]);
