
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