'use strict';

var fs = require('fs'),
	path = require('path');

function getDirectories(srcpath) {
  var arr = [];
  fs.readdirSync(srcpath).filter(function(file) {
  	if(fs.statSync(path.join(srcpath, file)).isDirectory() && file.indexOf('.') !== 0){
  		arr.push({
  			srcpath: srcpath,
  			file: file,
  			stat: fs.statSync(path.join(srcpath, file))
  		});
  	}
  });
  return arr;
}
function getUserHome() {
  return process.env.HOME || process.env.HOMEPATH || process.env.USERPROFILE;
}

module.exports = {
	getUserHome: function(){
		return getDirectories(getUserHome());
	}
}