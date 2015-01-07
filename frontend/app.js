'use strict';

// Declare app level module which depends on views, and components
angular.module('workend', [
  'ngMaterial',
  'ngRoute',
  'ngStorage'
]).run(['$rootScope', '$http', 'WEsession', '$sessionStorage', function ($rootScope, $http, WEsession, $sessionStorage) {
    $rootScope.$watch('currentUser', function(currentUser) {
      $rootScope.currentUser = $sessionStorage.currentUser ? $sessionStorage.currentUser : $rootScope.currentUser;
      if (!$rootScope.currentUser) {
          WEsession.checkUser();
      }
    });

  }]).config(['$locationProvider', '$routeProvider', function($locationProvider, $routeProvider){
  	$locationProvider.html5Mode({
  		enabled: true,
  		requireBase: false
  	});
  	$routeProvider
      .when('/', {
      	templateUrl: '/api/v1/partials/index',
        controller: 'webtop'
      })
      .when('/login', {
        templateUrl: '/api/v1/partials/login',
        controller: 'login'
      })
      .when('/signup', {
        templateUrl: '/api/v1/partials/login',
        controller: 'login'
      })
      //===== mean-cli hook =====//
      .otherwise({
        redirectTo: '/'
      });
  }]);