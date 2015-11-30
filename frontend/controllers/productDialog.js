'use strict';

angular.module('workend').controller('productDialog', ['$scope', '$mdDialog', '$http', '$location', '$mdToast', function($scope, $mdDialog, $http, $location, $mdToast){
	$scope.currentContent = [];
	$scope.oldPath = '';
	$scope.select = function(event, name){
		event.preventDefault();
		$scope.new.product.path = $scope.oldPath + '/' + name;
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
	$scope.close = function(){
		$mdDialog.hide();
	};
	$scope.createNewProduct= function(){
		$http.post('product/new', $scope.new.product).success(function(data, status){
			if(status === 200){
				if($scope.$root.products){
					$scope.$root.products.push(data);
				} else {
					$scope.$root.products = [];
				}
				$scope.$root.prod = data;
				$scope.$root.sprint = { _id: '', name: 'Select Sprint'};
				$mdToast.show($mdToast.simple().content('Product created successfully'));
				$scope.$root.$broadcast('closeToolbar');
				$scope.close();
			}
		});
	}
	$scope.new = {
		product: {
			name: '',
			description: '',
			private: false,
			path: ''
		}
	}
	$scope.selectPath = function(item){
		$scope.selectedFolder = item;
	};
	$scope.projects = [];
	$scope.selectedFolder = {};
	$http.get('/api/v1/projects/')
			.success(function (response, status, headers, config) {
				$scope.dirloaded = true;
				$scope.projects = response;
			})
			.error(function(error, status, headers, config) {
			  	// $location.path('/login');
				$mdToast.show($mdToast.simple().content('Please login or sign up'));
			});
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
			  	// $location.path('/login');
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