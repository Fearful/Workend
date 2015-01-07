var express = require('express');
var path = require('path');
var favicon = require('serve-favicon');
var logger = require('morgan');
var cookieParser = require('cookie-parser');
var bodyParser = require('body-parser');
var passport = require('passport');
var session = require('express-session');
var MongoStore = require('connect-mongo')(session);
var GITHUB_CLIENT_ID = "dd86fb0cf3aebd2d3b27"
var GITHUB_CLIENT_SECRET = "4a56925b1d45917fd9cc5a3a24b51fa717d0b204";
var GitHubStrategy = require('passport-github').Strategy;
var LocalStrategy = require('passport-local').Strategy;

var mongoose = require('mongoose');
var dbName = 'workend';
var connectionString = 'mongodb://localhost:27017/' + dbName;
mongoose.connect(connectionString);
var app = express();

var User = require(__dirname + '/backend/models/user');

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

// passport.use(new GitHubStrategy({
//     clientID: GITHUB_CLIENT_ID,
//     clientSecret: GITHUB_CLIENT_SECRET,
//     callbackURL: "http://127.0.0.1:3000/auth/github/callback"
//   },
//   function(accessToken, refreshToken, profile, done) {
//     // asynchronous verification, for effect...
//     process.nextTick(function () {
      
//       // To keep the example simple, the user's GitHub profile is returned to
//       // represent the logged-in user.  In a typical application, you would want
//       // to associate the GitHub account with a user record in your database,
//       // and return that user instead.
//       return done(null, profile);
//     });
//   }
// ));

var routes = require('./backend/routes/index');

// view engine setup
app.set('views', path.join(__dirname, 'frontend'));
app.set('view engine', 'jade');

// uncomment after placing your favicon in /public
//app.use(favicon(__dirname + '/public/favicon.ico'));
app.use(logger('dev'));
app.use(bodyParser.json());
app.use(bodyParser.urlencoded({ extended: false }));
app.use(cookieParser());
app.use(session({
    secret: 'workend',
    store: new MongoStore({ url: 'mongodb://localhost/workend', cookie: { maxAge: 60000 }}),
    resave: true,
    db: mongoose.connection.db,
    saveUninitialized: true
}));
app.use(passport.initialize());
app.use(passport.session());
app.use(express.static(path.join(__dirname, 'frontend')));

app.use('/', routes);
app.use('/*', function(req, res){
  res.render('index', { title: 'Workend', user: req.user });
});
// catch 404 and forward to error handler
app.use(function(req, res, next) {
    var err = new Error('Not Found');
    err.status = 404;
    next(err);
});

// error handlers

// development error handler
// will print stacktrace
if (app.get('env') === 'development') {
    app.use(function(err, req, res, next) {
        res.status(err.status || 500);
        res.render('error', {
            message: err.message,
            error: err
        });
    });
}

// production error handler
// no stacktraces leaked to user
app.use(function(err, req, res, next) {
    res.status(err.status || 500);
    res.render('error', {
        message: err.message,
        error: {}
    });
});

var server = app.listen(5000, function () {

  var host = server.address().address
  var port = server.address().port

  console.log('Workend app listening on localhost at port', port)

});

module.exports = app;
