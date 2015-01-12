var mongoose = require('mongoose')
   ,Schema = mongoose.Schema;
 
var projectSchema = new Schema({
    dateAdded: {type: Date, default: Date.now},
    author: {type: String, default: ''},
    description: {type: String, default: ''},
    mainFile: {type: String, default: ''},
    version: {type: String, default: '0.0.0'},
    name: {type: String, default: ''},
    owner: {type: String},
    path: {type: String}
});
 
module.exports = mongoose.model('Project', projectSchema, 'projects');