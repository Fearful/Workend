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
	getDirs: function (req, res, next) {
      var path = req.params.path;
      if(path){
        res.json(path);
      } else {
        res.json(getDirectories(getUserHome()));
      }
  }
}