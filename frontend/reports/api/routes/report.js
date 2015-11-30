var express = require('express'),
	router = express.Router();

//
// Report C.R.U.D.
//
router.post('/report/new', ensureAuthenticated, function (req, res){
	//Create a new report
});

router.get('/report/:reportId', ensureAuthenticated, function (req, res){
	//Read report
});

router.put('/report/:reportId', ensureAuthenticated, function (req, res){
	//Update report
});

router.delete('/report/:reportId', ensureAuthenticated, function (req, res){
	//Delete report
});
//
// Report methods (related to the report model)
//
router.post('/report/config', ensureAuthenticated, function (req, res){
	//Configure personal widgets
});

router.post('/report/configOutput', ensureAuthenticated, function (req, res){
	//Configure output reports
});

router.post('/report/scheduleReport', ensureAuthenticated, function (req, res){
	//Schedule a report for a date instance
});

module.exports = router;

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}
