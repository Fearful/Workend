'use strict';

var mongoose = require('mongoose'),
    passport = require('passport'),
    github = require('octonode');
exports.session = function (req, res) {
    if(!req.user){
        //302 redirect
        res.redirect('/login');
        return;
    }
    if(req.user.gitToken){
        var client = github.client(req.user.gitToken);
        client.get('/user/repos', {}, function (err, status, body, headers) {
            res.json({user: req.user, githubRepos: body});
        });
    } else {
        res.json({ user: req.user });
    }
};
exports.logout = function (req, res) {
    if (req.user) {
        req.logout();
        res.sendStatus(200)
    } else {
        res.sendStatus(400, 'Not logged in');
    }
};
exports.login = function (req, res, next) {
    passport.authenticate('local', function (err, user, info) {
        var error = err || info;
        if (error) {
            return res.status(400).json(error);
        }
        req.logIn(user, function (err) {
            if (err) {
                return res.send(err);
            }
            if(req.user.gitToken){
                var client = github.client(req.user.gitToken);
                client.get('/user/repos', {}, function (err, status, body, headers) {
                    res.json({ user: req.user, githubRepos: body });
                });
            } else {
                res.json({ user: req.user });
            }
        });
    })(req, res);
}