var express = require('express');
var router = express.Router();
var session = require('../controllers/session');
//
// Session C.R.D.
//
router.post('/', session.login);

router.get('/', session.session);

router.delete('/', ensureAuthenticated, session.logout);

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}

module.exports = router;