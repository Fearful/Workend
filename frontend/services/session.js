'use strict';

angular.module('workend').service('WEsession', ['$rootScope', '$location', '$http', '$mdToast', function($rootScope, $location, $http, $mdToast){
	var service = {};

	service.login = function(username, password){
		$http.post('session', { username: username, password: password }).
			success(function(data, status, headers, config) {
			$rootScope.savedProduct = data.user.saved_product;
			$rootScope.savedProject = data.user.saved_project;
			$rootScope.currentUser = data.user;
			window.location.reload();
		}).error(function(data, status, headers, config) {
				$mdToast.show($mdToast.simple().content(data.message));
			});
	};

	service.logout = function(){
		$http.delete('session');
		$rootScope.currentUser = false;
		$mdToast.show($mdToast.simple().content('Logout successfully!'));
		window.location.reload();
	};

	service.register = function(user, email, pass, companyName, companyDescription){
		$http.post('user/new', { username: user, email: email, password: pass, companyName: companyName, companyDescription: companyDescription }).
			success(function(data, status, headers, config) {
				$rootScope.currentUser = data;
				$mdToast.show($mdToast.simple().content('Registration complete!'));
				$location.path('/newUser')
			}).
			error(function(data, status, headers, config) {
				$mdToast.show($mdToast.simple().content(data.message));
			});
	};

	service.checkUser = function(cb){
		$http.get('session')
			.success(function (response, status, headers, config) {
				$rootScope.currentUser = response.user;
				if(!cb){
					return;
				}
				cb(response);
			})
			.error(function(error, status, headers, config) {
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
				$location.path('/login');
			});
	};

	return service;
}]);