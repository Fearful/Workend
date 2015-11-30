'use strict';
//Main server file
var express = require('express'),
    path = require('path'),
    logger = require('morgan'),
    cookieParser = require('cookie-parser'),
    bodyParser = require('body-parser'),
    compression = require('compression'),
    passport = require('passport'),
    session = require('express-session'),
    MongoStore = require('connect-mongo')(session),
    serveIndex = require('serve-index'),
    responseTime = require('response-time'),
    errorhandler = require('errorhandler'),
    timeout = require('connect-timeout'),
    favicon = require('serve-favicon'),
    serveStatic = require('serve-static'),
    http = require('http'),
    https = require('https'),
    port = 5000,
    router = express.Router();
var mongoose = require('mongoose');
mongoose.connect('mongodb://localhost:27017/workend');

var app = express();
app.use(favicon(__dirname + '/frontend/assets/img/favicon.ico'))
app.use(timeout('10s'));
app.use(responseTime())
app.use(compression())
app.use(bodyParser.json({limit: '25mb'}));
app.use(bodyParser.urlencoded({ extended: false, limit: '25mb'}));
var server = http.createServer(app)
var io = require('socket.io')(server);
require('./backend/config/passport');
app.use(logger('dev'));
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
app.use('/static', serveStatic(path.join(__dirname,'frontend')));

//Require index of routes and setup API
//This creates the api based on the file name, so that product.js will contain all urls inside /product/**/*
require('./backend/routes/index')(app);
//Prototype api
require('./frontend/pr0t0/api/routes/index')(app);
//Boards api
require('./frontend/boards/api/routes/index')(app);

//Serve file system
app.use('/api/fs', serveIndex(getUserHome(), {'icons': true, 'template': 'backend/templates/files.html', 'view': 'details'}));

//Render tool specific partials
app.use('**/partials/:partial', function (req, res){
   var partial = req.params.partial;
   if(req.isAuthenticated()){
      res.render('.' + req.params[0] + '/partials/'+partial+'.jade');
   } else {
      res.render('./partials/login.jade');
   }
});
//Render workend partials
app.use('partials/:partial', function (req, res){
   var partial = req.params.partial;
   if(req.isAuthenticated()){
      res.render('./partials/'+partial+'.jade');
   } else {
      res.render('./partials/login.jade');
   }
});
//Render index
app.use('/', function (req, res, next) {
    if(req.params.length > 0) next();
    res.render('./layout', { title: 'Workend', user: req.user });
});

// view engine setup
app.set('views', path.join(__dirname, 'frontend'));
app.set('view engine', 'jade');

// catch 404 and forward to error handler
app.use(function(req, res, next) {
    var err = new Error('Not Found');
    err.status = 404;
    next(err);
});
// development error handler
// will print stacktrace
if (app.get('env') === 'development') {
    app.use(errorhandler());
}

// production error handler
// no stacktraces leaked to user
// app.use(function(err, req, res, next) {
//     res.status(err.status || 500);
//     res.render('error', {
//         message: err.message,
//         error: {}
//     });
// });
server.listen(port, function () {
    console.log('Workend app listening on localhost at port %d', port);
});
//On connection of client
io.on('connection', function (client){ 
    client.on('itemChanged', function(item) {
        client.broadcast.emit('updateItem', item);
    });
    client.on('itemDeleted', function(item){
        client.broadcast.emit('deleteItem', item);
    });
    client.on('addItem', function(item){
        client.broadcast.emit('itemCreated', item);
    });
  // new client is here!
  // console.log('client connected');
});
function getUserHome() {
  return process.env.HOME || process.env.HOMEPATH || process.env.USERPROFILE;
}

module.exports = app;