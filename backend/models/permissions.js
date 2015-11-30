var mongoose = require('mongoose')
   ,Schema = mongoose.Schema;

var permissions_schema = new Schema({
    name : String,
    title : String
});

var roles_schema = new Schema({
    name : {type: String},
    title : {type: String},
    userId: {type: String},
    _permissions : { type:Array }
});

var roles_settings = new Schema({
    // ?
});

module.exports = mongoose.model('Role', roles_schema);
module.exports = mongoose.model('Permission', permissions_schema);