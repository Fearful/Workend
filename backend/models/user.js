'use strict';

var mongoose = require('mongoose'),
  Schema = mongoose.Schema,
  crypto = require('crypto');
 
var UserSchema = new Schema({
  email: { type: String, unique: true, required: true },
  username: { type: String, unique: true, required: true },
  hashedPassword: { type:String },
  salt: { type:String },
  name: { type:String },
  avatar: { type: String, default: '/static/assets/uploads/userImages/' + this._id + '.png' },
  github: { type:Object },
  provider: { type:String },
  saved_product: { type: String },
  saved_project: { type: String },
  preferredLanguage: { type: String }
});
/**
 * Virtuals
 */
UserSchema.virtual('password').set(function(password) {
    this._password = password;
    this.salt = this.makeSalt();
    this.hashedPassword = this.encryptPassword(password);
  }).get(function() {
    return this._password;
  });
UserSchema.virtual('user_info').get(function () {
    return { '_id': this._id, 'username': this.username, 'email': this.email, 'avatar': this.avatar };
});
/**
 * Validations
 */
var validatePresenceOf = function (value) {
  return value && value.length;
};
UserSchema.path('email').validate(function (email) {
  var emailRegex = /^([\w-\.]+@([\w-]+\.)+[\w-]{2,4})?$/;
  return emailRegex.test(email);
}, 'The specified email is invalid.');
/**
 * Pre-save hook
 */
UserSchema.pre('save', function(next) {
  if (!this.isNew) {
    return next();
  }

  if (!validatePresenceOf(this.password)) {
    next(new Error('Invalid password'));
  }
  else {
    next();
  }
});
/**
 * Methods
 */
UserSchema.methods = {
  authenticate : function(plainText) {
    return this.encryptPassword(plainText) === this.hashedPassword;
  },
  makeSalt : function() {
    return crypto.randomBytes(16).toString('base64');
  },
  encryptPassword : function(password) {
    if (!password || !this.salt) return '';
    var salt = new Buffer(this.salt, 'base64');
    return crypto.pbkdf2Sync(password, salt, 10000, 64).toString('base64');
  }
};
module.exports = mongoose.model('User', UserSchema, 'users');