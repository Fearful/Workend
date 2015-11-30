var mongoose = require('mongoose'),
	Products = mongoose.model('Product'),
	Workspace = mongoose.model('Workspace'),
	User = mongoose.model('User'),
	ObjectId = mongoose.Types.ObjectId,
	_ = require('underscore');

exports.create = function (req, res, next) {
	req.body.owner = req.user._id;
	req.body.created_by = req.user._id;
	req.body.modified_by = req.user._id;
	// NEED TO add a minimalistic userlist with avatar, name, and id 
	req.body.members = [{
		_id: req.user._id,
		avatar: '',
		name: req.user.username,
		role: 'owner'
	}];
  var newProduct = new Products(req.body);
  newProduct.save(function(err) {
    if (err) {
      return res.status(400).json(err);
    }
    Workspace.find({created_by: req.user._id}, function(err, workspace){
		workspace[0].products.push(newProduct._id);
		workspace[0].save(function(err) {
		    if (err) {
		      return res.status(400).json(err);
		    }
		    newProduct.workspace = workspace[0]._id;
		    newProduct.save(function(err){
		    	if(err){
		    		return res.status(400).json(err);
		    	} else {
		    		return res.status(200).json(newProduct);
		    	}
		    });
		});
	});
  });

};
exports.read = function (req, res, next) {
	var projectId = req.params.projectId;
	if(projectId){
		Products.findById(ObjectId(projectId), function (err, project) {
		    if (err) {
		      return next(new Error('Failed to load Product'));
		    }
		    if (project) {
		      res.send(project);
		    } else {
		      res.send(404, 'PROYECT_NOT_FOUND')
		    }
		});
	} else {
		Products.find(function (err, project) {
		    if (err) {
		      return next(new Error('Failed to load Product'));
		    }
		    if (project) {
		      res.send(project);
		    } else {
		      res.send(404, 'PROYECT_NOT_FOUND')
		    }
		});
	}
};
exports.addUser = function(req, res, next){
	var productId = req.body.productId;
	var searchedUser = false;
	User.find({email: req.body.email }, function(err, user){
		if(err){
			return next(new Error('Failed to find user'));
		}
		if(user){
			searchedUser = user[0];
		}
	});
	Products.findById(ObjectId(productId), function (err, product) {
	    if (err) {
	      return next(new Error('Failed to add member to product'));
	    }
	    if (product) {
			product.members.push({
				_id: searchedUser._id,
				avatar: searchedUser.avatar,
				name: searchedUser.username,
				role: req.body.role
			});
			product.save(function(err){
				if(err){
					res.status(400).send('COUDNT_SAVE_PROJECT');
				}
				res.status(200).send(product.dev_list);
			});
	    } else {
	    	res.status(400).send('CANNOT_ADD_MEMBER');
	    }
	});
}
exports.update = function (req, res, next) {
	var projectId = req.params.projectId;
	Products.findById(ObjectId(projectId), function (err, project) {
	    if (err) {
	      return next(new Error('Failed to load Product'));
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
	Products.remove({ _id: projectId }, function(err, movie) {
	    if (err) {
	      return res.send(err);
	    }
	    res.json({ message: 'Successfully deleted' });
	});
};
