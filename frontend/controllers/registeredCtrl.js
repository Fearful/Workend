'use strict';

angular.module('workend').controller('registeredCtrl', ['$window', '$scope', '$http', '$location', function($window, $scope, $http, $location){
	$scope.completed = false;
	$scope.picture = false;
	$scope.formPicture = false;
	$scope.$watch('picture', function(newVal, oldVal){
		if(newVal && newVal != oldVal){
			$http.post('upload/image', { image: newVal, id: $scope.$root.currentUser._id })
			.success(function(data){
				debugger;
				$scope.completed = true;
			});
		}
	});
	$scope.continue = function(){
		$location.path('/');
	}
}]);