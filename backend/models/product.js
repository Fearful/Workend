var mongoose = require('mongoose')
   ,Schema = mongoose.Schema,
   _ = require('underscore');
 
var productSchema = new Schema({
    private: {type: Boolean, default: false},
    name: {type: String, default: ''},
    description: {type: String, default: ''},
    created_at: {type: Date, default: Date.now},
    created_by: {type: String},
    modified_at: {type:Date, default:Date.now},
    modified_by: {type: String},
    sprints: {type:Array},
    documents: {type: Array, default: []},
    members: { type: Array, default: []},
    owner: { type: String },
    current_version: { type: Number },
    workspace: { type: String },
    path: { type: String },
    project_status: { type: Array, default: [ { name: 'To Do', value: 0 }, { name: 'In Progress', value: 1 }, { name: 'In review', value: 2 }, { name: 'Closed', value: 3 } ] }
}, { minimize: false });
productSchema.set('toJSON', { getters: true, virtuals: true });
productSchema.virtual('devs').get(function () {
  return filtered_members = _.filter(this.members, function(mem){
    return mem.role === 'Developer';
  });
});
productSchema.virtual('qas').get(function (name) {
  return filtered_members = _.filter(this.members, function(mem){
    return mem.role === 'QA';
  });
});
productSchema.virtual('analist').get(function (name) {
  return filtered_members = _.filter(this.members, function(mem){
    return mem.role === 'Analist';
  });
});
productSchema.virtual('product_owner').get(function (name) {
  return filtered_members = _.filter(this.members, function(mem){
    return mem.role === 'Product Owner';
  });
});
productSchema.virtual('scrum_master').get(function (name) {
  return filtered_members = _.filter(this.members, function(mem){
    return mem.role === 'Scrum Master';
  });
});
 
module.exports = mongoose.model('Product', productSchema, 'products');