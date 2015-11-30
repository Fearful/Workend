var fs = require('fs'),
	walk = require('walk'),
  	path = require('path'),
	sys = require('sys'),
	exec = require('child_process').exec;
exports.run = function(req, res, next){
	var projectPath = req.body.pathToProject;
	var task = req.body.task;
	var framework = req.body.framework;
	projectPath = getUserHome() + '/' + projectPath.substring(projectPath.indexOf('/api/fs') + 8);
	if(projectPath && task && framework){
		if(framework === 'gulp'){
			var process = exec('gulp ' + task, {
			  cwd: projectPath
			}, function(error, stdout, stderr){
			    if (error !== null) {
			      console.log('exec error: ' + error);
			  	}
			});
		} else if(framework === 'grunt'){
			var process = exec('grunt ' + task, {
			  cwd: projectPath
			}, function(error, stdout, stderr){
				console.log('stdout: ' + stdout);
			    console.log('stderr: ' + stderr);
			    if (error !== null) {
			      console.log('exec error: ' + error);
			  	}
			});
		}
		process.on('close', function (code) {
		  console.log('child process exited with code ' + code);
		});
		return res.status(200);
	} else {
		res.sendStatus(400, 'No task or framework sended');
	}
};

function getUserHome() {
  return process.env.HOME || process.env.HOMEPATH || process.env.USERPROFILE;
}