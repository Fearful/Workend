var express = require('express');
var router = express.Router();
var jade = require('jade');
var passport = require('passport');
var github = require('octonode');

//
// Timesheet C.R.U.D.
//
router.post('/timesheet/new', ensureAuthenticated, function (req, res){
	//Create a new timesheet
});

router.get('/timesheet/:timesheetId', ensureAuthenticated, function (req, res){
	//Read timesheet
});

router.put('/timesheet/:timesheetId', ensureAuthenticated, function (req, res){
	//Update timesheet
});

router.delete('/timesheet/:timesheetId', ensureAuthenticated, function (req, res){
	//Delete timesheet
});
//
// Timesheet methods (related to the timesheet model)
//
router.get('/timesheet/addhours', ensureAuthenticated, function (req, res){
	//Add hours to a specific date instance
});

module.exports = router;

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}
