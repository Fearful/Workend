var mongoose = require('mongoose')
   ,Schema = mongoose.Schema;
 
var projectSchema = new Schema({
    created_at: {type: Date, default: Date.now},
    created_by: {type: String, default: ''},
    description: {type: String, default: ''},
    mainFile: {type: String, default: ''},
    current_version: {type: String, default: '0.0.0'},
    name: {type: String, default: ''},
    path: {type: String},
    status: {type: String, default: 'OPEN'},
    requirement_sprint: { type: Boolean, default: false },
    product_id: { type: String },
    start_date: { type: Date },
    end_date: { type: Date },
    iteration: { type: Number }
});
 
module.exports = mongoose.model('Project', projectSchema, 'projects');