var express = require('express');
var router = express.Router();
var jade = require('jade');
var passport = require('passport');
var github = require('octonode');

//
// Document C.R.U.D.
//
router.post('/document/new', ensureAuthenticated, function (req, res){
	//Create a new document
});

router.get('/document/:documentId', ensureAuthenticated, function (req, res){
	//Read document
});

router.put('/document/:documentId', ensureAuthenticated, function (req, res){
	//Update document
});

router.delete('/document/:documentId', ensureAuthenticated, function (req, res){
	//Delete document
});
//
// Document methods (related to the document model)
//
router.get('/document/getDocuments', ensureAuthenticated, function (req, res){
	//Get recent documents modified or created by the current user
});

module.exports = router;

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}
