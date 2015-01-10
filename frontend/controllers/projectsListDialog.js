'use strict';

angular.module('workend').controller('projectsListDialog', ['$scope', '$mdDialog', '$mdSidenav', '$http', '$location', '$mdToast', function($scope, $mdDialog, $mdSidenav, $http, $location, $mdToast){
	$scope.projects = [];
	$scope.selectedProject = '';
	$scope.open = function(item){
		if(!item){
			var item = {};
			item._id = '';
			angular.forEach($scope.projects, function(proj){
				if(proj.name === $scope.selectedProject){
					item._id = proj._id;
				}
			})
		}
		$http.get('/api/v1/projects/' + item._id)
			.success(function (response, status, headers, config) {
				$scope.dirloaded = true;
				$mdToast.show($mdToast.simple().content('Opening Project...'));
				$mdDialog.hide();
				$mdSidenav('left').toggle();
				debugger;
			})
			.error(function(error, status, headers, config) {
			  	$location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
	};
	$scope.close = function(name){
		$mdDialog.hide();
	};
	$scope.selectProject = function(item){
		$scope.selectedProject = item.name;
	};
	$scope.dirloaded = false;
	if($scope.$root.currentUser){
		$http.get('/api/v1/projects/')
			.success(function (response, status, headers, config) {
				$scope.dirloaded = true;
				$scope.projects = response;
			})
			.error(function(error, status, headers, config) {
			  	$location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
	}
}]);