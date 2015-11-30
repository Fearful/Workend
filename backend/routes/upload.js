var express = require('express');
var router = express.Router();
var multer = require('multer');
var path = require('path');
var upload = multer({ dest: './frontend/assets/uploads/'});
//
// 
//
router.post('/image', ensureAuthenticated, function(req,res){
	var base64Data = req.body.image.replace(/^data:image\/png;base64,/, "");
	require("fs").writeFile(path.resolve(__dirname, '../../frontend/assets/uploads/userImages/', req.body.id + ".png"), base64Data, 'base64', function(err) {
	  if(err) res.sendStatus(400, 'Cant save image');
	  if(!err) res.sendStatus(200);
	});

    // upload(req.body,res,function(err) {
    //     if(err) {
    //         return res.end("Error uploading file.");
    //     }
    //     res.end("File is uploaded");
    // });
});

function ensureAuthenticated(req, res, next) {
  if (req.isAuthenticated()) { return next(); }
  res.render('./partials/login.jade');
}

module.exports = router;