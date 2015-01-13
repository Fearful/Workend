var fs = require('fs'),
	walk = require('walk'),
  	path = require('path'),
	sys = require('sys'),
	exec = require('child_process').exec;

//Path to ignore
var excludedPaths = [
  	'node_modules'
];
// File types
var types = {
	svg: 'Images',
	png: 'Images',
	jpg: 'Images',
	jpeg: 'Images',
	gif: 'Images',
	ico: 'Images',
	eot: 'Fonts',
	ttf: 'Fonts',
	woff: 'Fonts',
	js: 'Javascript',
	coffee: 'Javascript',
	rb: 'Ruby',
	erb: 'Ruby',
	py: 'Python',
	sh: 'Shell',
	java: 'Java',
	php: 'Php',
	c: 'C',
	h: 'C',
	cpp: 'C++',
	cc: 'C++',
	cxx: 'C++',
	pl: 'Perl',
	pm: 'Perl',
	t: 'Perl',
	ep: 'Perl',
	m: 'Objective-C',
	mm: 'Objective-C',
	asm: 'Assembly',
	clj: 'Clojure',
	go: 'Go',
	lisp: 'Lisp',
	hs: 'Haskell',
	pde: 'Processing',
	scm: 'Scheme',
	html: 'Markup',
	htm: 'Markup',
	xhtml: 'Markup',
	xml: 'Markup',
	css: 'Styles',
	mustache: 'Templates',
	haml: 'Templates',
	jade: 'Templates',
	less: 'Preprocessed CSS',
	sass: 'Preprocessed CSS',
	styl: 'Preprocessed CSS',
	md: 'Docs',
	markdown: 'Docs',
	cfg: 'Settings',
	ini: 'Settings',
	json: 'Serialized',
	yml: 'Serialized',
	DS_Store: 'Settings',
	bowerrc: 'Settings',
	gitignore: 'Settings',
	jshintrc: 'Settings',
};
exports.getStadistics = function(req, res, next){
	var path = req.body.path;
	path = getUserHome() + '/' + path.substring(path.indexOf('/api/fs') + 8);
	if(fs.existsSync(path + '/package.json')){
		var pkg = fs.readFileSync(path + '/package.json', {encoding: 'utf8'});
	}
	if(!pkg){
		res.sendStatus(400, 'No package.json Found');
	}
	if(fs.existsSync(path + '/gulpfile.js')){
		var gulp = fs.readFileSync(path + '/gulpfile.js', {encoding: 'utf8'});
		if(gulp){
			var tasks = [];
			while(gulp.indexOf('gulp.task(') !== -1){
				var index = gulp.indexOf('gulp.task(');
				var substring = gulp.substr(index + 11);
				var singleQuote = substring.indexOf("'");
				var doubleQuote = substring.indexOf('"');
				if(singleQuote != -1 && (singleQuote < doubleQuote || doubleQuote === -1)){
					tasks.push({ name: substring.substr(0, singleQuote), icon: 'gulp' })
					gulp = gulp.substr(index + singleQuote);
				} else if(doubleQuote != -1 && (doubleQuote < singleQuote || singleQuote === -1)){
					tasks.push({ name: substring.substr(0, doubleQuote), icon: 'gulp' })
					gulp = gulp.substr(index + doubleQuote);
				}
			}
		}
	}
	if(fs.existsSync(path + '/Gruntfile.js')){
		var grunt = fs.readFileSync(path + '/Gruntfile.js', {encoding: 'utf8'});
		if(grunt){
			var tasks = [];
			while(grunt.indexOf('grunt.registerTask(') !== -1){
				var index = grunt.indexOf('grunt.registerTask(');
				var substring = grunt.substr(index + 20);
				var singleQuote = substring.indexOf("'");
				var doubleQuote = substring.indexOf('"');
				if(singleQuote != -1 && (singleQuote < doubleQuote || doubleQuote === -1)){
					tasks.push({ name: substring.substr(0, singleQuote), icon: 'grunt' })
					grunt = grunt.substr(index + singleQuote);
				} else if(doubleQuote != -1 && (doubleQuote < singleQuote || singleQuote === -1)){
					tasks.push({ name: substring.substr(0, doubleQuote), icon: 'grunt' })
					grunt = grunt.substr(index + doubleQuote);
				}
			}
		}
	}
	if(fs.existsSync(path + '/.gitignore')){
			var ignore = fs.readFileSync(path + '/.gitignore', {encoding: 'utf8'});
		if(ignore){
			ignore = ignore.split('\n');
			for (var i = ignore.length - 1; i >= 0; i--) {
				ignore[i] = path + '/' + ignore[i]
			};
			ignore.push(path + '/.git');
		}
		var results = getAllFiles(path, ignore);
	} else {
		var results = getAllFiles(path, [path + '/.git']);
	}
	if(pkg){
		results.package = pkg;
	}
	if(gulp || grunt){
		results.tasks = tasks;
	}
	res.send(results);
};
function getAllFiles(relativePath, ignoredPaths){
	var num = 0;
	var files = {};
	var lines = {};
	var walker = walk.walkSync(relativePath, { filters: ignoredPaths, listeners:{
		names: function (root, nodeNamesArray) {
	        nodeNamesArray.sort(function (a, b) {
	          if (a > b) return 1;
	          if (a < b) return -1;
	          return 0;
	        });
	      },
		file: function(root, fileStats, next){
			var ext = fileStats.name.substring(fileStats.name.lastIndexOf('.') + 1);
			var lang = types[ext];
			if(typeof lang === 'string'){
				if(files[lang]){
					files[lang].count = files[lang].count + 1;
				} else {
					files[lang] = {lang: lang, count: 1};
				}
				if(lang !== 'Images' && lang !== 'Fonts'){
					if(lines[lang]){
						lines[lang].count = lines[lang].count + getFileLines(path.resolve(root, fileStats.name));
					} else {
						lines[lang] = { lang: lang, count: getFileLines(path.resolve(root, fileStats.name))};
					}
				}
			} else {
				if(files.Other){
					files.Other.count = files.Other.count + 1;
				} else {
					files.Other = {lang: 'Other', count: 1};
				}
			}
			next();
		}
	} });

	return { fileCount: files, lineCount: lines };
}

function getFileLines(path){
	var num = 0
	var ext = path.substring(path.lastIndexOf('.') + 1);
	var lang = types[ext];
	if(lang){
		var file = fs.readFileSync(path, {encoding: 'utf8'});
		file = file.split(/\r\n|[\n\r\u0085\u2028\u2029]/g);
		num = file.length;
	}
	return num;
}
function getUserHome() {
  return process.env.HOME || process.env.HOMEPATH || process.env.USERPROFILE;
}
