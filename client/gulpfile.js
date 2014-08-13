var browserify = require('browserify');
var concat = require('gulp-concat');
var gulp = require('gulp');
var jshint = require('gulp-jshint');
var minifyCSS = require('gulp-minify-css');
var source = require('vinyl-source-stream');
var streamify = require('gulp-streamify');
var uglify = require('gulp-uglify');

gulp.task('jshint', function() {
    gulp.src(['./gulpfile.js',
              './src/**/*.js'])
        .pipe(jshint())
        .pipe(jshint.reporter('default'));
});

gulp.task('build-js', function() {
    browserify({
        entries: [
            './index.js',
            './node_modules/bootstrap/dist/js/bootstrap.js'
        ],
        debug: true
    })
        .bundle()
        .pipe(source('build.js'))
        // .pipe(streamify(uglify())) // TODO: Figure out why this breaks
        .pipe(gulp.dest('.'));
});

gulp.task('build-css', function() {
        gulp.src([
            './node_modules/bootstrap/dist/css/bootstrap.css',
            './node_modules/rickshaw/rickshaw.css',
            './src/**/*.css'
        ])
        .pipe(minifyCSS({root: '.'}))
        .pipe(concat('build.css'))
        .pipe(gulp.dest('.'));
});

gulp.task('watch', function() {
    gulp.watch(['./src/**/*'], ['default']);
});

gulp.task('default', [
    'jshint',
    'build-js',
    'build-css'
]);
