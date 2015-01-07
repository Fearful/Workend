'use strict';

angular.module('workend').service('WEsession', ['$rootScope', '$sessionStorage', '$location', '$http', '$timeout', '$mdToast', function($rootScope, $sessionStorage, $location, $http, $timeout, $mdToast){
	var service = {};

	service.login = function(username, password){
		$http.post('/auth/session', { username: username, password: password }).
			success(function(data, status, headers, config) {
			$rootScope.currentUser = data;
			$location.path('/');
			$timeout(function(){
				$rootScope.$apply();
				$mdToast.show($mdToast.simple().content('Login successfully!'));
			});
			}).
			error(function(data, status, headers, config) {
				$mdToast.show($mdToast.simple().content(data.message));
			});
	};

	service.logout = function(){
		if($sessionStorage.currentUser){
			$sessionStorage.$reset();
		}
		$http.delete('/auth/session/');
		$location.path('/login');
		$timeout(function(){
			$rootScope.$apply();
			$mdToast.show($mdToast.simple().content('Logout successfully!'));
		}, 0);
	};

	service.register = function(user, email, pass){
		$http.post('/auth/users', { username: user, email: email, password: pass }).
			success(function(data, status, headers, config) {
			$rootScope.currentUser = data;
			$location.path('/');
			$timeout(function(){
				$rootScope.$apply();
				$mdToast.show($mdToast.simple().content('Registration complete!'));
			}, 0);
			}).
			error(function(data, status, headers, config) {
				$mdToast.show($mdToast.simple().content(data.message));
			});
	};

	service.checkUser = function(){
		$http.get('/auth/session/')
			.success(function (response, status, headers, config) {
				$sessionStorage.currentUser = response;
				$mdToast.show($mdToast.simple().content('Welcome back ' + response.username + '!'));
			})
			.error(function(error, status, headers, config) {
			  $location.path('/login');
			  $timeout(function(){
				$rootScope.$apply();
				$mdToast.show($mdToast.simple().content('Error while checking your user!'));
			  }, 0);
			});
	};

	return service;
}]);