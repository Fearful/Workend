var express = require('express'),
 path = require('path'),
 favicon = require('serve-favicon'),
 logger = require('morgan'),
 cookieParser = require('cookie-parser'),
 bodyParser = require('body-parser'),
 passport = require('passport'),
 session = require('express-session'),
 MongoStore = require('connect-mongo')(session),
 serveIndex = require('serve-index');

var mongoose = require('mongoose');
mongoose.connect('mongodb://localhost:27017/workend');

var app = express();

require('./backend/config/passport');

var routes = require('./backend/routes/index');

// view engine setup
app.set('views', path.join(__dirname, 'frontend'));
app.set('view engine', 'jade');
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
// Angular needs to handle the routing of the application
app.use('/login', function(req, res){
  res.render('index', { title: 'Workend', user: req.user });
});

//Serve file system
app.use('/api/fs', serveIndex(getUserHome(), {'icons': true, 'template': 'backend/templates/files.html', 'view': 'details'}));

// catch 404 and forward to error handler
app.use(function(req, res, next) {
    var err = new Error('Not Found');
    err.status = 404;
    next(err);
});

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

function getUserHome() {
  return process.env.HOME || process.env.HOMEPATH || process.env.USERPROFILE;
}

module.exports = app;