var mongoose = require('mongoose'),
	Products = mongoose.model('Product'),
	Projects = mongoose.model('Project'),
	Workspace = mongoose.model('Workspace'),
	ObjectId = mongoose.Types.ObjectId;

exports.getProducts = function(req, res, next){
	Products.find({'members._id': {$in: [req.user._id]}}, function(err, products){
		if (err) {
		    return res.status(400).json(err);
		}
		for (var i = products.length - 1; i >= 0; i--) {
			products[i].sprints = getProjects(products[i]._id, products.length, i, function(){
				res.status(200).json(products);
			});
		};
	});
}
function getProjects(productId, total, index, cb){
	Projects.find({product_id: productId}, function(err, projects){
		if(!err){
			return projects;
		} else {
			return [];
		}
	});	
	if(total - 1 === index){
		cb();
	}
}
// exports.create = function (req, res, next) {
// 	req.body.owner = req.user._id;
// 	req.body.created_by = req.user._id;
// 	req.body.modified_by = req.user._id;
// 	// NEED TO add a minimalistic userlist with avatar, name, and id 
// 	req.body.members = [];
//   var newProduct = new Products(req.body);
//   newProduct.save(function(err) {
//     if (err) {
//       return res.status(400).json(err);
//     }
//     Workspace.find({created_by: req.user._id}, function(err, workspace){
// 		workspace[0].products.push(newProduct._id);
// 		workspace[0].save(function(err) {
// 		    if (err) {
// 		      return res.status(400).json(err);
// 		    }
// 		    return res.status(200).json(newProduct);
// 		});
// 	});
//   });

// };
// exports.read = function (req, res, next) {
// 	var projectId = req.params.projectId;
// 	if(projectId){
// 		Products.findById(ObjectId(projectId), function (err, project) {
// 		    if (err) {
// 		      return next(new Error('Failed to load Product'));
// 		    }
// 		    if (project) {
// 		      res.send(project);
// 		    } else {
// 		      res.send(404, 'PROYECT_NOT_FOUND')
// 		    }
// 		});
// 	} else {
// 		Products.find(function (err, project) {
// 		    if (err) {
// 		      return next(new Error('Failed to load Product'));
// 		    }
// 		    if (project) {
// 		      res.send(project);
// 		    } else {
// 		      res.send(404, 'PROYECT_NOT_FOUND')
// 		    }
// 		});
// 	}
// };
// exports.update = function (req, res, next) {
// 	var projectId = req.params.projectId;
// 	Products.findById(ObjectId(projectId), function (err, project) {
// 	    if (err) {
// 	      return next(new Error('Failed to load Product'));
// 	    }
// 	    if (project) {
// 	      res.send(project);
// 	    } else {
// 	      res.send(404, 'PROYECT_NOT_FOUND')
// 	    }
// 	});
// };
// exports.delete = function (req, res, next) {
// 	var projectId = req.params.projectId;
// 	Products.remove({ _id: projectId }, function(err, movie) {
// 	    if (err) {
// 	      return res.send(err);
// 	    }
// 	    res.json({ message: 'Successfully deleted' });
// 	});
// };
