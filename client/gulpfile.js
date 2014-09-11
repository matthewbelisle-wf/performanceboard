var browserify = require('browserify');
var concat = require('gulp-concat');
var gulp = require('gulp');
var jshint = require('gulp-jshint');
var minifyCSS = require('gulp-minify-css');
var source = require('vinyl-source-stream');
var streamify = require('gulp-streamify');
var uglify = require('gulp-uglify');

var BUILD_DIR = '../app/static';

gulp.task('jshint', function() {
    gulp.src(['./gulpfile.js',
              './src/**/*.js'])
        .pipe(jshint())
        .pipe(jshint.reporter('default'));
});

gulp.task('build-copy', function() {
    return gulp.src([
        // TODO: bootstrap
        './img/**/*'
    ], {base: '.'})
        .pipe(gulp.dest(BUILD_DIR));
});

gulp.task('build-js', function() {
    browserify({
        entries: './src/main.js',
        debug: true
    })
        .bundle()
        .pipe(source('build.js'))
        // .pipe(streamify(uglify())) // TODO: Figure out why this breaks
        .pipe(gulp.dest(BUILD_DIR));
});

gulp.task('build-css', function() {
        gulp.src([
            './node_modules/bootstrap/dist/css/bootstrap.css',
            './node_modules/rickshaw/rickshaw.css',
            './src/**/*.css'
        ])
        .pipe(minifyCSS({root: BUILD_DIR}))
        .pipe(concat('build.css'))
        .pipe(gulp.dest(BUILD_DIR));
});

gulp.task('watch', function() {
    gulp.watch(['./src/**/*'], ['default']);
});

gulp.task('default', [
    'jshint',
    'build-js',
    'build-css',
    'build-copy'
]);
