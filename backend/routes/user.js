var express = require('express');
var router = express.Router();
var users = require('../controllers/users');
//
// User C.R.U.D.
//
router.post('/new', users.new);

router.get('/:userId', ensureAuthenticated, users.find);

router.put('/:userId', ensureAuthenticated, function (req, res){
	//Update user
});

router.delete('/:userId', ensureAuthenticated, function (req, res){
	//Delete user
});
//
// User methods (related to the user model)
//
router.post('/saveSession', ensureAuthenticated, users.saveSession);

router.get('/getProducts', ensureAuthenticated, function (req, res){
	//Get recent products and projects modified, assigned or owned by the current user
});

router.get('check_username/:username', users.check_username);

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}

module.exports = router;