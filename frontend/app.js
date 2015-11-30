'use strict';
var socket = io('http://localhost:5000');
  socket.connect();
  socket.on('news', function (data) {
    socket.emit('connect', 'test');
  });
// Declare app level module which depends on views, and components
angular.module('workend', [
  'ngMaterial',
  'ngRoute',
  'as.sortable',
  'pascalprecht.translate',
  'ngStorage',
  'camera',
  'tc.chartjs',
  'pr0t0Editor'
]).config(['$locationProvider', '$routeProvider', '$mdThemingProvider', '$translateProvider', function($locationProvider, $routeProvider, $mdThemingProvider, $translateProvider){
  	$locationProvider.html5Mode({
  		enabled: true,
  		requireBase: false
  	});
  	$routeProvider
    .when('/', {
      	templateUrl: '/partials/index',
        controller: 'workendCtrl'
      })
    .when('/boards', {
      templateUrl: 'boards/partials/index',
      controller: 'boardCtrl',
    })
    .when('/timesheet', {
      templateUrl: 'timesheet/partials/index',
      controller: 'timesheetCtrl',
    })
    .when('/calendar', {
      templateUrl: 'calendar/partials/index',
      controller: 'calendarCtrl',
    })
    .when('/pr0t0', {
      templateUrl: 'pr0t0/partials/index',
      controller: 'pr0t0Controller',
    })
    .when('/login', {
      templateUrl: '/partials/login',
      controller: 'login'
    })
    .when('/newUser', {
      templateUrl: '/partials/registered',
      controller: 'registeredCtrl'
    });
    $mdThemingProvider.theme('default')
      .primaryPalette('light-blue')
      .accentPalette('indigo');

    //Languages
    // $translateProvider.translations('en', {
    //   HEADLINE: 'Hello there, This is my awesome app!',
    //   INTRO_TEXT: 'And it has i18n support!'
    // })
    // .translations('es', {
    //   HEADLINE: 'Hello there, This is my awesome app!',
    //   INTRO_TEXT: 'And it has i18n support!'
    // });
    $translateProvider.useStaticFilesLoader({
        prefix: '/static/assets/lang/locale-',
        suffix: '.json'
    });
    // $translateProvider.preferredLanguage('es_ES');
    // $translateProvider.preferredLanguage('en_US');
    $translateProvider.fallbackLanguage('en_US');
    $translateProvider.determinePreferredLanguage(function () {
      var lang = window.navigator.userLanguage || window.navigator.language;
      return lang;
    });
    // $translateProvider.rememberLanguage(true);
  }]);