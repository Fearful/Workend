'use strict';

// Declare app level module which depends on views, and components
angular.module('workend', [
  'ngMaterial',
  'ngRoute',
  'ngStorage',
  'tc.chartjs'
]).run(['$rootScope', '$http', 'WEsession', '$sessionStorage', '$location', '$mdToast', function ($rootScope, $http, WEsession, $sessionStorage, $location, $mdToast) {
    $rootScope.$watch('currentUser', function(currentUser) {
      $rootScope.currentUser = $sessionStorage.currentUser ? $sessionStorage.currentUser : $rootScope.currentUser;
      if (!$rootScope.currentUser && $location.path() !== '/login') {
          WEsession.checkUser(function(response){
            $rootScope.userLoaded = true;
            $mdToast.show($mdToast.simple().content('Welcome back ' + response.username + '!'));
          });
      } else if($location.path() !== '/login'){
        $rootScope.userLoaded = true;
        $mdToast.show($mdToast.simple().content('Welcome back ' + $rootScope.currentUser.username + '!'));
      }
    });

  }]).config(['$locationProvider', '$routeProvider', '$mdThemingProvider', function($locationProvider, $routeProvider, $mdThemingProvider){
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
      .otherwise({
        redirectTo: '/'
      });
      $mdThemingProvider.theme('default')
        .primaryPalette('light-blue')
        .accentPalette('deep-purple');
  }]);