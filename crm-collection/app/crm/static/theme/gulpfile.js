var gulp = require('gulp');
var sass = require('gulp-sass');
var rename = require('gulp-rename');
var cleanCSS = require('gulp-clean-css');

//style paths
var globalSCSS = 'sass/global/*.scss',
    globalDest = 'assets/global/css/',
    adminSCSS = 'sass/admin/*.scss/layout/',
    adminDest = 'assets/global/layout/css/';

gulp.task('global', function(){
    gulp.src(globalSCSS)
        .pipe(sass().on('error', sass.logError))
        .pipe(gulp.dest(globalDest));
});

gulp.task('global_min', function(){
    gulp.src(globalDest + "/*.css")
        .pipe(cleanCSS({compatibility: 'ie8'}))
        .pipe(rename({suffix: '.min'}))
        .pipe(gulp.dest(globalDest));
});

gulp.task('admin', function(){
    gulp.src(adminSCSS)
        .pipe(sass().on('error', sass.logError))
        .pipe(gulp.dest(adminDest))
});

gulp.task('admin_min', function(){
    gulp.src(adminDest + "/*.css")
        .pipe(cleanCSS({compatibility: 'ie8'}))
        .pipe(rename({suffix: '.min'}))
        .pipe(gulp.dest(adminDest));
});

gulp.task('watch',function() {
    gulp.watch(globalSCSS,['global','global_min','admin','admin_min']);
});

