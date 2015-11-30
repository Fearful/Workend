'use strict';

var mongoose = require('mongoose'),
  User = mongoose.model('User'),
  Workspace = mongoose.model('Workspace'),
  passport = require('passport'),
  ObjectId = mongoose.Types.ObjectId;
exports.new = function (req, res, next) {
  var newUser = new User(req.body);
  newUser.provider = 'local';

  newUser.save(function(err) {
    if (err) {
      return res.status(400).json(err);
    }
    req.logIn(newUser, function(err) {
      if (err) return next(err);
      var newWorkspace = new Workspace({
        created_by: newUser.user_info._id,
        modified_by: newUser.user_info._id,
        name: req.body.companyName || '',
        description: req.body.companyDescription || ''
      })
      newWorkspace.save(function(err) {
        if (err) {
          return res.status(400).json(err);
        }
        return res.status(200).json(newUser.user_info, newWorkspace);
      });
    });
  });
};
exports.find = function (req, res, next) {
  var userId = req.params.userId;
  User.findById(ObjectId(userId), function (err, user) {
    if (err) {
      return next(new Error('Failed to load User'));
    }
    if (user) {
      res.send({username: user.username, profile: user.profile });
    } else {
      res.send(404, 'USER_NOT_FOUND')
    }
  });
};
exports.check_username = function (req, res, next) {
  var username = req.params.username;
  User.findOne({ username : username }, function (err, user) {
    if (err) {
      return next(new Error('Failed to load User ' + username));
    }

    if(user) {
      res.json({exists: true});
    } else {
      res.json({exists: false});
    }
  });
}
exports.saveSession = function(req, res, next){
  var update = {};
  if(req.body.preferredLanguage){
    update.preferredLanguage = req.body.preferredLanguage;
  } else {
    update.saved_product = req.body.productId;
    update.saved_project = req.body.projectId;
  }
  User.update({ _id: ObjectId(req.user._id)}, update, function(err, user){
    if (err) {
      return res.status(400).json(err);
    }
    return res.status(200).json(user);
  });
}