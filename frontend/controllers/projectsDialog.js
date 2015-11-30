'use strict';

angular.module('workend').controller('projectsDialog', ['$scope', '$mdDialog', '$http', '$location', '$mdToast', function($scope, $mdDialog, $http, $location, $mdToast){
	$scope.close = function(name){
		$mdDialog.hide();
	};
	$scope.createNewProject= function(){
		$http.post('project/new', $scope.new.project).success(function(data, status){
			if(status === 200){
				$scope.$root.prod.sprints.push(data);
				$scope.$root.sprint = data;
				$mdToast.show($mdToast.simple().content('Sprint created successfully'));
				$scope.$root.$broadcast('closeToolbar');
				$scope.close();
			}
		});
	}
	$scope.selectProject = function(name){
		$scope.dirloaded = false;
		if($scope.$root.currentUser){
			$scope.$root.currentUser.starred = $scope.oldPath + '/' + $scope.new.project.path;
		}
		$http.post('/api/v1/projects/new', { name: $scope.selectedFolder, owner: $scope.$root.currentUser._id, path: $scope.oldPath + '/' + $scope.selectedFolder })
			.success(function (response, status, headers, config) {
				$scope.dirloaded = true;
				$mdDialog.hide();
				$mdToast.show($mdToast.simple().content('Project added successfully'));
				//Once the project is added we need to close the dialog and open the project in the view
			})
			.error(function(error, status, headers, config) {
			  	// $location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
	};
	$scope.new = {
		project: {
			name: '',
			description: '',
			backlog: false,
			requirement_sprint: false,
			start_date: new Date(),
			end_date: $scope.$root.prod.end_date || new Date,
			product_id: $scope.$root.prod._id,
			path: ''
		}
	}
}]);