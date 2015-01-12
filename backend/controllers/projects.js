var mongoose = require('mongoose'),
	Projects = mongoose.model('Project'),
	ObjectId = mongoose.Types.ObjectId;

exports.create = function (req, res, next) {
  var newProject = new Projects(req.body);
  newProject.save(function(err) {
    if (err) {
      return res.status(400).json(err);
    }
    return res.status(200).json(newProject);
  });
};
exports.read = function (req, res, next) {
	var projectId = req.params.projectId;
	if(projectId){
		Projects.findById(ObjectId(projectId), function (err, project) {
		    if (err) {
		      return next(new Error('Failed to load Project'));
		    }
		    if (project) {
		      res.send(project);
		    } else {
		      res.send(404, 'PROYECT_NOT_FOUND')
		    }
		});
	} else {
		Projects.find(function (err, project) {
		    if (err) {
		      return next(new Error('Failed to load Project'));
		    }
		    if (project) {
		      res.send(project);
		    } else {
		      res.send(404, 'PROYECT_NOT_FOUND')
		    }
		});
	}
};
exports.update = function (req, res, next) {
	var projectId = req.params.projectId;
	Projects.findById(ObjectId(projectId), function (err, project) {
	    if (err) {
	      return next(new Error('Failed to load Project'));
	    }
	    if (project) {
	      res.send(project);
	    } else {
	      res.send(404, 'PROYECT_NOT_FOUND')
	    }
	});
};
exports.delete = function (req, res, next) {
	var projectId = req.params.projectId;
	Projects.remove({ _id: projectId }, function(err, movie) {
	    if (err) {
	      return res.send(err);
	    }
	    res.json({ message: 'Successfully deleted' });
	});
};
