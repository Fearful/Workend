'use strict';

var mongoose = require('mongoose'),
    passport = require('passport');

exports.session = function (req, res) {
    if(!req.user){
        res.sendStatus(400, 'Not logged in');
        return;
    }
    res.json(req.user.user_info);
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
            res.json(req.user.user_info);
        });
    })(req, res, next);
}