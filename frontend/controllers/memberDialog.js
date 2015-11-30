'use strict';

angular.module('workend').controller('memberDialog', ['$scope', '$mdDialog', '$http', '$location', '$mdToast', function($scope, $mdDialog, $http, $location, $mdToast){
	$scope.close = function(){
		$mdDialog.hide();
	};
	$scope.product = $scope.$root.prod;
	$scope.addNewUser= function(){
		$http.post('product/member', $scope.new.user).success(function(data, status){
			if(status === 200){
				$mdToast.show($mdToast.simple().content('User added successfully'));
				$scope.$root.$broadcast('closeToolbar');
				$scope.close();
			}
		});
	}
	$scope.new = {
		user: {
			email: '',
			role: '',
			productId: $scope.$root.prod._id
		}
	}
	$scope.roles = [ 'Developer', 'QA', 'Admin' ];
	$scope.prodRoles = [ 'Developer', 'QA', 'Admin', 'owner' ];
}]);