'use strict';

angular.module('workend').controller('login', ['$window', '$scope', 'WEsession', function($window, $scope, WEsession){
	$scope.$root.username = '';
	$scope.$root.email = '';
	$scope.$root.password = '';
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
}]);