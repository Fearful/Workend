var express = require('express');
var router = express.Router();
var jade = require('jade');
var passport = require('passport');

/* GET home page. */
router.get('/', function(req, res) {
  res.render('index', { title: 'Workend', user: req.user });
});

// Github routes
// router.get('/auth/github',
//   passport.authenticate('github'),
//   function(req, res){
//   });
// router.get('/auth/github/callback', 
//   passport.authenticate('github', { failureRedirect: '/login' }),
//   function(req, res) {
//     res.redirect('/');
//   });

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
