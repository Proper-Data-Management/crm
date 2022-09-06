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
        "module": "kz.gov.pki.knca.commonUtils",
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
            data.method = "getKeyInfo";
            data.args = [storageAlias];
            getEDSData();
            ncaPassword = "";
        }
    };

    /** signXml() - подписываем XML
     *====================================================================*/
    var signXml = function (xmlData) {

		console.log("edsKeys",edsKeys);
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
                "SIGNATURE",
                xmlData,
                '',
                ''
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
                "SIGNATURE",
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

        console.log("data method: " + data.method);

        if(str == heartbeatMsg) {
            return;
        }

        answer = JSON.parse(str);

        if (answer['code'] === "500") {
            alert(answer['message']);
        } else if (answer['code'] === "200") {
            switch (data.method) {
                case 'getKeyInfo':
                    NCALayer.postGetSubjectDN(answer.responseObject);
                    break;
                case 'signXml' :
                    NCALayer.postSignXml(answer.responseObject);
                    NCALayer.showPerson(false);
                    break;
                case 'showFileChooser' :
                    NCALayer.postShowFileChooser(answer.responseObject);
                    break;
                case 'createCMSSignatureFromFile' :
                    NCALayer.postCreateCMSSignatureFromFile(answer.responseObject);
                    NCALayer.showPerson(false);
                    break;
            }
        } else {
            NCALayer.showError(answer);
        }
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
        $theScope.signXml = signXml;
        $theScope.showFileChooser = showFileChooser;
        $theScope.createCMSSignatureFromFile = createCMSSignatureFromFile;
    };

    return NCALayer;
}]);
