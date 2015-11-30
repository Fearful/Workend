var mongoose = require('mongoose')
   ,Schema = mongoose.Schema;
 
var workspaceSchema = new Schema({
    created_at: {type: Date, default: Date.now},
    created_by: {type: Object},
    modified_at: {type:Date, default:Date.now},
    modified_by: {type: Object},
    name: {type: String, default: ''},
    description: {type: String, default: ''},
    products: {type: Array, default: []},
    allow_dev: {type: Object, default: {
        create_branches: true,
        schedule_own_tasks: true,
        allow_deploy: true,
        create_documents: true,
        create_events: true,
        create_tasks: true,
        access_tests: true,
    }},
    allow_qa: {type: Object, default: {
        create_branches: false,
        schedule_own_tasks: true,
        allow_deploy: true,
        create_documents: true,
        create_events: false,
        create_tasks: true,
        access_tests: true,
    }}
});
 
module.exports = mongoose.model('Workspace', workspaceSchema, 'workspaces');