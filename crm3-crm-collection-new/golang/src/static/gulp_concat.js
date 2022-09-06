var concat = require('gulp-concat');
var gulp = require('gulp');
var minify = require('gulp-minify');

var sizereport = require('gulp-sizereport');


gulp.task('scripts', function() {
  return gulp.src([
  
'theme/assets/global/plugins/respond.min.js',
'theme/assets/global/plugins/excanvas.min.js',
'js/plugins/jquery/jquery-2.0.2.min.js',
//'js/plugins/moment/moment-with-locales.min.js',
'js/plugins/moment/moment.min.js',
'js/plugins/moment/ru.js',
'theme/assets/global/plugins/jquery-migrate.min.js',
'theme/assets/global/plugins/bootstrap/js/bootstrap.min.js',
'theme/assets/global/plugins/bootstrap-hover-dropdown/bootstrap-hover-dropdown.min.js',
'theme/assets/global/plugins/jquery-slimscroll/jquery.slimscroll.min.js',
'theme/assets/global/plugins/jquery.blockui.min.js',
'theme/assets/global/plugins/jquery.cokie.min.js',
'theme/assets/global/plugins/uniform/jquery.uniform.min.js',
'theme/assets/global/plugins/bootstrap-switch/js/bootstrap-switch.min.js',
'theme/assets/global/plugins/angularjs/angular.min.js',
'theme/assets/global/plugins/angularjs/angular-sanitize.min.js',
'theme/assets/global/plugins/angularjs/angular-touch.min.js',
'theme/assets/global/plugins/angularjs/plugins/angular-ui-router.min.js',
'theme/assets/global/plugins/angularjs/plugins/ocLazyLoad.min.js',
'theme/assets/global/plugins/angularjs/plugins/ui-bootstrap-tpls-2.5.0.min.js',
'js/i18n/angular-locale_kk.js',
'js/tmhDynamicLocale.js',
'js/angular-file-upload.min.js',
'js/plugins/angular-tree-control/angular-tree-control.js',
'js/angular-pubsub.min.js',
'js/plugins/ui-select/select.min.js',
'js/plugins/chat.js/Chart.bundle.min.js',
'js/plugins/chat.js/angular-chart.js',
'js/plugins/webcam/webcam.min.js',
'js/plugins/webcam/ng-webcam.js',
'js/ng-map.min.js',
'js/plugins/moment/angular-moment.min.js',
'js/app.js',
'js/components.js',
'js/directives.js',
'theme/assets/global/scripts/metronic.min.js',
'theme/assets/admin/layout/scripts/layout.js',
'theme/assets/admin/layout/scripts/quick-sidebar.js',
'theme/assets/admin/layout/scripts/demo.js',
'theme/assets/admin/pages/scripts/index.js',
'theme/assets/admin/pages/scripts/tasks.js',
'js/plugins/rendro-easy-pie-chart/easypiechart.min.js',
'js/plugins/rendro-easy-pie-chart/jquery.easypiechart.js',
'js/plugins/rendro-easy-pie-chart/angular.easypiechart.min.js',
'js/services/services-restapi.js',
'js/services/deal-services.js',
'js/services/services-ui-simpletable.js',
'js/services/services-dml.js',
'js/services/services-ui.js',
'js/mask.min.js',
'js/angular-bootstrap-checkbox.js',
'js/controllers/SimpleTableController.js',
'js/controllers/SimpleDetailTableController.js',
'js/services/services-eds.js',
'js/angular-translate.min.js',
'js/smart-table.min.js',
'js/plugins/flowchart/svg_class.js',
'js/plugins/flowchart/mouse_capture_service.js',
'js/plugins/flowchart/dragging_service.js',
'js/plugins/flowchart/flowchart_viewmodel.js',
'js/plugins/flowchart/flowchart_directive.js',
'js/plugins/angular-css-injector/angular-css-injector.min.js',
'js/plugins/ace/ui-ace.js',
'js/plugins/ace/ext-language_tools.js',
'js/plugins/x2js/xml2json.min.js',
'js/plugins/x2js/json2xml.js',
'js/plugins/angular-ui-tree/angular-ui-tree.min.js',
'js/plugins/angular-bootstrap-colorpicker/js/bootstrap-colorpicker-module.min.js',
'js/plugins/ckeditor/angular-ckeditor.min.js',
'js/prototype.js'
  ])
	.pipe(sizereport())
    .pipe(concat('all.js'))
    .pipe(gulp.dest('./js/'));
});


gulp.task('compress1', function() {
  gulp.src('./theme/assets/global/scripts/metronic.js')
    .pipe(minify({
        ext:{
            src:'-debug.js',
            min:'.min.js'
        },
        exclude: ['tasks'],
        ignoreFiles: ['.combo.js', '-min.js', '.min.js']
    }))
    .pipe(gulp.dest('./theme/assets/global/scripts/'))
});


gulp.task('compress2', function() {
  gulp.src('./bpmn.io/bower_components/bpmn-js/dist/bpmn-modeler.js')
    .pipe(minify({
        ext:{
            src:'-debug.js',
            min:'.min.js'
        },
        exclude: ['tasks'],
        ignoreFiles: ['.combo.js', '-min.js']
    }))
    .pipe(gulp.dest('./bpmn.io/bower_components/bpmn-js/dist/'))
});

gulp.start( 'compress1');
gulp.start( 'compress2');
gulp.start( 'scripts');
