'use strict';

angular.module('workend').controller('projectsDialog', ['$scope', '$mdDialog', '$http', '$location', '$mdToast', function($scope, $mdDialog, $http, $location, $mdToast){
	$scope.currentContent = [];
	$scope.oldPath = '';
	$scope.select = function(event, name){
		event.preventDefault();
		$scope.selectedFolder = name;
	};
	$scope.open = function(event, name){
		event.preventDefault();
		if(!name){
			var backDir = $scope.oldPath.split('/');
			backDir.pop();
			backDir = backDir.join('/');
			goTo(backDir);
			return;
		}
		goTo($scope.oldPath + '/' + name);
	};
	$scope.close = function(name){
		$mdDialog.hide();
	};
	$scope.selectProject = function(name){
		$scope.dirloaded = false;
		if($scope.$root.currentUser){
			$scope.$root.currentUser.starred = $scope.oldPath + '/' + $scope.selectedFolder;
		}
		$http.post('/api/v1/projects/new', { name: $scope.selectedFolder, owner: $scope.$root.currentUser._id, path: $scope.oldPath + '/' + $scope.selectedFolder })
			.success(function (response, status, headers, config) {
				$scope.dirloaded = true;
				$mdDialog.hide();
				$mdToast.show($mdToast.simple().content('Project added successfully'));
				//Once the project is added we need to close the dialog and open the project in the view
			})
			.error(function(error, status, headers, config) {
			  	$location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
	};
	$scope.dirloaded = false;
	if($scope.$root.currentUser){
		goTo('api/fs');
	}
	function goTo(path){
		$scope.dirloaded = false;
		$http.get(path)
			.success(function (response, status, headers, config) {
				$scope.dirloaded = true;
				$scope.oldPath = path;
				$scope.currentContent = response;
			})
			.error(function(error, status, headers, config) {
			  	$location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
	}
	$scope.data = {
      selectedIndex : 0
    };
    $scope.next = function() {
      $scope.data.selectedIndex = Math.min($scope.data.selectedIndex + 1, 2) ;
    };
    $scope.previous = function() {
      $scope.data.selectedIndex = Math.max($scope.data.selectedIndex - 1, 0);
    };
}]);