var express = require('express');
var router = express.Router();
var projects = require('../controllers/projects');
//
// Project C.R.U.D.
//
router.post('/new', ensureAuthenticated, projects.create);

router.get('/:projectId', ensureAuthenticated, projects.read);

router.put('/:projectId', ensureAuthenticated, projects.update);

router.delete('/:projectId', ensureAuthenticated, projects.delete);
//
// Project methods (related to the project model)
//
router.post('/import', ensureAuthenticated, function (req, res){
	//Import an xsl file to convert into db task and issues, provided configuration rules
});


function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}

module.exports = router;