var express = require('express');
var router = express.Router();
var jade = require('jade');
var passport = require('passport');
var github = require('octonode');

/* GET home page. */
router.get('/', function(req, res) {
	res.render('index', { title: 'Workend', user: req.user });
});

// API jade partials
router.get('/api/v1/partials/:partial', function (req, res){
   var partial = req.params.partial
   if(req.isAuthenticated()){
      res.render('./partials/'+partial+'.jade');
   } else {
      res.render('./partials/login.jade');
   }
});
var fs = require('../controllers/fs');
router.get('/api/v1/navigate/:path', fs.getDirs);
// User Routes
var Users = require('../controllers/users');
router.post('/auth/users', Users.create);
router.get('/auth/users/:userId', Users.show);
router.get('/auth/check_username/:username', Users.exists);

// Session Routes
var session = require('../controllers/session');
router.get('/auth/session', session.session);
router.post('/auth/session', session.login);
router.delete('/auth/session', session.logout);

// Github routes
router.get('/auth/github',
  passport.authorize('github'));

router.get('/auth/github/callback', 
  passport.authorize('github', { failureRedirect: '/auth/github/fail' }),
  function(req, res) {
    // Successful authentication, redirect home.
    res.redirect('/');
  });
router.get('/auth/github/fail', 
	function(req, res){
		res.redirect('/');
	});
router.get('github/repos',
	function(req, res){
		var client = github.client(req.user.gitToken);
		client.get('/user/repos', {}, function (err, status, body, headers) {
		    res.json(body);
		});
	});
// Project routes
var Projects = require('../controllers/projects.js');
router.post('/api/v1/projects/new', ensureAuthenticated, Projects.create);
router.get('/api/v1/projects/', ensureAuthenticated, Projects.read);
router.get('/api/v1/projects/:projectId', ensureAuthenticated, Projects.read);
router.put('/api/v1/projects/:projectId', ensureAuthenticated, Projects.update);
router.delete('/api/v1/projects/:projectId', ensureAuthenticated, Projects.delete);

// Stadistics
var Stadistics = require('../controllers/codeStadistics.js');
router.post('/api/v1/statistics', ensureAuthenticated, Stadistics.getStadistics);

// Tasks
var Tasks = require('../controllers/tasks.js');
router.post('/api/v1/tasks', ensureAuthenticated, Tasks.run);

module.exports = router;

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.redirect('/login')
}
