'use strict';

angular.module('workend').controller('login', ['$window', '$scope', 'WEsession', '$location', function($window, $scope, WEsession, $location){
	$scope.$root.username = '';
	$scope.$root.email = '';
	$scope.$root.password = '';
	$scope.$root.userLoaded = true;
	$scope.login = function(){
		WEsession.login($scope.$root.username, $scope.$root.password)
	};
	$scope.register = function(){
		WEsession.register($scope.$root.username, $scope.$root.email, $scope.$root.password);
	};
	$scope.data = {
      selectedIndex : 0
    };
    $scope.next = function() {
      $scope.data.selectedIndex = Math.min($scope.data.selectedIndex + 1, 2) ;
    };
    $scope.previous = function() {
      $scope.data.selectedIndex = Math.max($scope.data.selectedIndex - 1, 0);
    };
    if($scope.$root.currentUser){
    	$location.path('/');
    }
}]);