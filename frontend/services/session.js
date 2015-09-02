'use strict';

angular.module('workend').service('WEsession', ['$rootScope', '$sessionStorage', '$location', '$http', '$mdToast', function($rootScope, $sessionStorage, $location, $http, $mdToast){
	var service = {};

	service.login = function(username, password){
		$http.post('/auth/session', { username: username, password: password }).
			success(function(data, status, headers, config) {
			data.user.starred = data.starredProject;
			$rootScope.currentUser = data.user;
			$location.path('/');
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
		$rootScope.currentUser = false;
		$mdToast.show($mdToast.simple().content('Logout successfully!'));
	};

	service.register = function(user, email, pass){
		$http.post('/auth/users', { username: user, email: email, password: pass }).
			success(function(data, status, headers, config) {
			$rootScope.currentUser = data;
			$location.path('/');
			$mdToast.show($mdToast.simple().content('Registration complete!'));
			}).
			error(function(data, status, headers, config) {
				$mdToast.show($mdToast.simple().content(data.message));
			});
	};

	service.checkUser = function(callback){
		$http.get('/auth/session/')
			.success(function (response, status, headers, config) {
				response.user.starred = response.starred;
				$sessionStorage.currentUser = response.user;
				callback(response);
			})
			.error(function(error, status, headers, config) {
			  	$location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
	};

	return service;
}]);