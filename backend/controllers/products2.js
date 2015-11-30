'use strict';

var mongoose = require('mongoose'),
    passport = require('passport'),
    Products = mongoose.model('Product'),
    Projects = mongoose.model('Project'),
    Workspace = mongoose.model('Workspace'),
    User = mongoose.model('User'),
    ObjectId = mongoose.Types.ObjectId,
    github = require('octonode');


// Product C.R.U.D

exports.new = function (req, res, next) {
	req.body.owner = req.user._id;
	req.body.created_by = req.user._id;
	req.body.modified_by = req.user._id;
	// NEED TO add a minimalistic userlist with avatar, name, and id 
	req.body.members = [{
		_id: req.user._id,
		avatar: req.user.avatar,
		name: req.user.username,
		role: req.body.roleName,
		roleId: req.body.roleId
	}];
  	var newProduct = new Products(req.body);
	Workspace.findOne({created_by: req.user._id}, function(err, workspace){
		workspace.products.push(newProduct._id);
		workspace.save(function(err) {
		    if (err) {
		      return res.status(400).json(err);
		    }
		});
		newProduct.workspace = workspace._id;
	    newProduct.save(function(err){
	    	if(err){
	    		return res.status(400).json(err);
	    	} else {
	    		return res.status(200).json(newProduct);
	    	}
	    });
	});

};

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
	    	if(product.owner !== req.user._id){
	    		res.status(400).send('You´re not the owner of this product.');
	    	}
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
exports.read = function (req, res, next){
	if(!req.params.productId){
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
	} else {
		Products.findOne({'_id': req.params.productId }, function(err, products){
			if (err) {
			    return res.status(400).json(err);
			}
			products.sprints = getProjects(products._id, function(){
				res.status(200).json(products);
			});

		});
	}
}
function getProjects(productId, total, index, cb){
	Projects.find({product_id: productId}, function(err, projects){
		if(!err){
			return projects;
		} else {
			return [];
		}
	});	
	if(typeof total === 'function'){
		total();
		return;
	}
	if(total - 1 === index){
		cb();
	}
}