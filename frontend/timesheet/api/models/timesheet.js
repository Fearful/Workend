var mongoose = require('mongoose')
   ,Schema = mongoose.Schema;
 
var timesheetSchema = new Schema({
    start_date: { type: Date, default: Date.now },
    end_date: { type: Date, default: Date.now },
    hours: { type: Array, default: [] },
    owner: { type: String },
    modified_by: { type: String },
    modified_at: { type: Date, default: Date.now },
    assigned_to: { type: Array, default: [] }
});
 
module.exports = mongoose.model('Timesheet', timesheetSchema, 'timesheets');