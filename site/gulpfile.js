var gulp = require('gulp');
var plugins = require('gulp-load-plugins')();

gulp.task('hugo:prod', function(callback) {
    gulp.src('').pipe(plugins.shell(['hugo -v -b https://www.chameth.com/ -d /var/www/html/'], { cwd: process.cwd() })).on('end', callback || function() {});;
});

gulp.task('default', ['hugo:prod']);