/*
 * Imports
 */
var gulp		= require('gulp');
var uglify		= require('gulp-uglify');
var concat		= require('gulp-concat');
var sass		= require('gulp-sass');
var plumber		= require('gulp-plumber');

/*
 * Overrides
 */
var src = gulp.src;
gulp.src = function() {
  return src.apply(gulp, arguments)
    .pipe(plumber(function(error) {
      util.log(util.colors.red(
        'Error (' + error.plugin + '): ' + error.message
      ));
      notifier.notify({
        title: 'Error (' + error.plugin + ')',
        message: error.message.split('\n')[0]
      });
      this.emit('end');
    })
  );
};

/*
 * Asset pipelines
 */

// SCSS
gulp.task('assets:scss', function() {
	return gulp.src('scss/**/*.scss')
		.pipe(sass({
			outputStyle: 'compressed',
			sourceMapEnabled: true,
			includePaths: [
				/* dependencies */
			]
		}).on('error', sass.logError))
		.pipe(gulp.dest('static/css/'));
});

/*
 * Commands / Interface
 */
gulp.task('default', ['assets:scss']);