'use strict';

angular.module('workend').service('WEsession', ['$rootScope', '$sessionStorage', '$location', '$http', function($rootScope, $sessionStorage, $location, $http){
	var service = {};

	service.login = function(username, password){
		$http.post('/auth/session', { username: username, password: password }).
			success(function(data, status, headers, config) {
			$rootScope.currentUser = data;
			$location.path('/');
			}).
			error(function(data, status, headers, config) {
			});
	};

	service.logout = function(){
		if($sessionStorage.currentUser){
			$sessionStorage.$reset();
		}
		$http.delete('/auth/session/');
		$location.path('/login');
	};

	service.register = function(user, email, pass){
		$http.post('/auth/users', { username: user, email: email, password: pass }).
			success(function(data, status, headers, config) {
			$rootScope.currentUser = data;
			$location.path('/');
			}).
			error(function(data, status, headers, config) {
			});
	};

	service.checkUser = function(){
		$http.get('/auth/session/')
			.success(function (response, status, headers, config) {
				$sessionStorage.currentUser = response;
			})
			.error(function(error, status, headers, config) {
			  $location.path('/login');
			  console.log(error)
			});
	};

	return service;
}]);