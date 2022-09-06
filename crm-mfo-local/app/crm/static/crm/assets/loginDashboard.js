window.bd = window.bd || {};
(function($, bd, undefined){

    var initLoginPage = function() {

        var $loginDashboard = $('.bd-login-dashboard'),
            $loginContainer = $('.bd-loginContainer-dashboard'),
            $loginForm = $('.bd-loginForm-dashboard');

        // Trigger login animations
        $loginDashboard.addClass('load');

        // Wait until fade animation ends
        $loginDashboard.on('transitionend', function(){
            var $loginError = $('.bd-login-error');

            if($loginError.length) {
                // Shake box if there is any login error
                $loginContainer.addClass('shake');
            }
        });



        // important. used to disable caching in back/forward (history) list
        // without this javascript will not run after history.back
        window.onunload = function(){};
    };

    // Preload images in background
    var preLoadImg = function(imgUrl, callback){
        var backgroundImg = document.createElement("img");

        backgroundImg.onload = function(){
            callback && callback();
        };
        backgroundImg.src = imgUrl;
    };

    bd.initLoginPage = initLoginPage;
    bd.preLoadImg = preLoadImg;

    // Show random background
    $(function(){

        var $loginError = $('.bd-login-error');

        if($.browser && $.browser.msie) {
            $('body').addClass('ie');
        }

        // TODO: get backgrounds from Content Services
        var backgrounds = [
            bd.contextRoot + 'assets/login-background1.jpg',
            bd.contextRoot + 'assets/login-background2.jpg',
            bd.contextRoot + 'assets/login-background3.jpg',
            bd.contextRoot + 'assets/login-background4.jpg',
            bd.contextRoot + 'assets/login-background5.jpg'
        ];

        // Get a random background
        var bg = backgrounds[Math.floor(Math.random()*backgrounds.length)];

        if (localStorage && !localStorage.getItem('bg')) {
            // Store background for first time
            localStorage.setItem('bg', bg);
        }

        if($loginError.length) {
            // Show previous background
            bg = localStorage.getItem('bg');
        } else {
            // Store background for next page loads (w/ errors)
            localStorage.setItem('bg', bg);
        }

        // Preload the background
        bd.preLoadImg(bg, function(){
            var $loginDashboardBg = $('.dashboard .bd-login-dashboard-bg'),
                $loginContainerBg = $('.dashboard .bd-loginContainer-dashboard-bg');

            $loginDashboardBg.css('background-image', 'url("' + bg + '")');
            $loginContainerBg.css('background-image', 'url("' + bg + '")');
        });
    });
})(jQuery, window.bd);

