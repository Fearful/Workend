var GITHUB_CLIENT_ID = "dd86fb0cf3aebd2d3b27"
var GITHUB_CLIENT_SECRET = "4a56925b1d45917fd9cc5a3a24b51fa717d0b204";
var GitHubStrategy = require('passport-github').Strategy;
var passport = require('passport');
var LocalStrategy = require('passport-local').Strategy;
var User = require('../models/user');
var Project = require('../models/project');
// Serialize sessions
passport.serializeUser(function(user, done) {
  done(null, user.id);
});

passport.deserializeUser(function(id, done) {
  User.findOne({ _id: id }, function (err, user) {
    done(err, user);
  });
});

// Use local strategy
passport.use(new LocalStrategy({
    usernameField: 'username',
    passwordField: 'password'
  },
  function(username, password, done) {
    User.findOne({ username: username }, function (err, user) {
      if (err) {
        return done(err);
      }
      if (!user) {
        return done(null, false, {
          'errors': {
            'username': { type: 'Username is not registered.' }
          }
        });
      }
      if (!user.authenticate(password)) {
        return done(null, false, {
          'errors': {
            'password': { type: 'Password is incorrect.' }
          }
        });
      }
      return done(null, user);
    });
  }
));

passport.use(new GitHubStrategy({
    clientID: GITHUB_CLIENT_ID,
    clientSecret: GITHUB_CLIENT_SECRET,
    callbackURL: "http://127.0.0.1:5000/auth/github/callback"
  },
  function(accessToken, refreshToken, profile, done) {
    // asynchronous verification, for effect...
    process.nextTick(function () {
      
      // To keep the example simple, the user's GitHub profile is returned to
      // represent the logged-in user.  In a typical application, you would want
      // to associate the GitHub account with a user record in your database,
      // and return that user instead.

      // User.findOne({ email: profile.emails[0].value }, function (err, user) {
      //   if(user) {
      //     user.github = profile;
      //     console.log(user)
      //     return done(null, profile);
      //   }
      // });

      User.update({email: profile.emails[0].value }, {
          githubId: profile.id,
          gitToken    : profile.token,
          gitName     : profile.displayName,
          github: profile,
      }, function(err, numberAffected, rawResponse) {
         //handle it
         console.log(numberAffected)
         console.log(rawResponse)
         return done(null, profile);
      });
    });
  }
));

module.exports = passport;