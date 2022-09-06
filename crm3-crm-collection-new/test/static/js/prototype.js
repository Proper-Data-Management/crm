
(function() {
    Date.prototype.toDBDateTime = Date_toDBDateTime;
    function Date_toDBDateTime() {
        var year, month, day;
        year = String(this.getFullYear());
        month = String(this.getMonth() + 1);
        hours = String(this.getHours());
        mins = String(this.getMinutes());
        secs = String(this.getSeconds());

        if (month.length == 1) {
            month = "0" + month;
        }
        if (hours.length == 1) {
            hours = "0" + hours;
        }
        if (mins.length == 1) {
            mins = "0" + mins;
        }
        if (secs.length == 1) {
            secs = "0" + secs;
        }

        day = String(this.getDate());
        if (day.length == 1) {
            day = "0" + day;
        }
        return year + "-" + month + "-" + day + " " + hours + ":" + mins + ":" + secs;
    }
})();
