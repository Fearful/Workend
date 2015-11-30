module.exports = function(config){
  config.set({

    basePath : '../frontend',

    files : [
      'socket.io/socket.io.js',
      'assets/lib/Chart.js/Chart.min.js',
      'assets/lib/angular/angular.min.js',
      'assets/lib/angular-aria/angular-aria.js',
      'assets/lib/angular-animate/angular-animate.js',
      'assets/lib/hammerjs/hammer.js',
      'assets/lib/angular-material/angular-material.min.js',
      'assets/lib/ngStorage/ngStorage.min.js',
      'assets/lib/angular-route/angular-route.min.js',
      'assets/lib/tc-angular-chartjs/dist/tc-angular-chartjs.min.js',
      'app.js',
      'controllers/webtop.js',
      'controllers/mainToolbar.js',
      'controllers/memberDialog.js',
      'controllers/login.js',
      'controllers/taskRunToaster.js',
      'boards/controllers/taskCreate.js',
      'controllers/projectsDialog',
      'controllers/productDialog',
      'services/session.js',
      'test/unit/**/*.js'

    ],

    autoWatch : true,

    frameworks: ['jasmine'],

    browsers : ['Chrome'],

    plugins : [
            'karma-chrome-launcher',
            'karma-firefox-launcher',
            'karma-jasmine',
            'karma-junit-reporter'
            ],

    junitReporter : {
      outputFile: 'test_out/unit.xml',
      suite: 'unit'
    }

  });
};