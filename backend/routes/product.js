var express = require('express');
var router = express.Router();

var products = require('../controllers/products2.js');
//
// Product C.R.U.D.
//
router.post('/new', ensureAuthenticated, products.new);

router.get('/:productId', ensureAuthenticated, products.read);

router.get('/', ensureAuthenticated, products.read);

router.put('/:productId', ensureAuthenticated, function (req, res){
	//Update product
});

router.delete('/:productId', ensureAuthenticated, function (req, res){
	//Delete product
});
//
// Product methods (related to the product model)
//
router.get('getReports/:productId', ensureAuthenticated, function (req, res){
	//Get all products and containing projects of the current user (using the req.user)
});

router.get('getStatistics/:productId', ensureAuthenticated, function (req, res){
	//Get a product current statistics considering branch
});

router.get('getCodeCoverage/:productId', ensureAuthenticated, function (req, res){
	//Get a product current code coverage considering branch
});

router.get('getTestsReport/:productId', ensureAuthenticated, function (req, res){
	//Get a product current test reports considering branch
});

router.get('readAvailableTests/:productId', ensureAuthenticated, function (req, res){
	//Get a product gruntjs and gulpjs tasks considering branch
});

router.post('/member', ensureAuthenticated, products.addUser);

router.put('member/:productId', ensureAuthenticated, function (req, res){
	//Remove member of a product
});

router.post('/role', ensureAuthenticated, function (req, res){
	//Add role to a product
});

router.put('role/:roleId', ensureAuthenticated, function (req, res){
	//Update role
});

router.delete('role/:roleId', ensureAuthenticated, function (req, res){
	//Delete role
});

router.get('getCurrentStatus/:productId', ensureAuthenticated, function (req, res){
	//Get current status of sprint and product via the product's projects
});

router.get('loadDebugTasks/:productId', ensureAuthenticated, function (req, res){
	//Retrieve any debug tasks founded inside the products folder
});

router.get('browseFs/:path', ensureAuthenticated, function (req, res){
	//Retrieve any debug tasks founded inside the products folder
});

router.get('loadRecentTests/:productId', ensureAuthenticated, function (req, res){
	//Load recent tests created
});

router.post('/scheduleTask', ensureAuthenticated, function (req, res){
	//Schedule tasks to run
});


function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}

module.exports = router;