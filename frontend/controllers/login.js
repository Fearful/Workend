'use strict';

angular.module('workend').controller('login', ['$window', '$scope', 'WEsession', '$location', '$route', function($window, $scope, WEsession, $location, $route){
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
	$scope.githubLogin = function(){
		$location.path('/auth/github');
	};
	$scope.data = {
      selectedIndex : 2
    };
    $scope.next = function() {
      $scope.data.selectedIndex = $scope.data.selectedIndex === 2 ? 0 : $scope.data.selectedIndex + 1;
    };
    $scope.previous = function() {
      $scope.data.selectedIndex = $scope.data.selectedIndex === 0 ? 2 : $scope.data.selectedIndex - 1;
    };
    if($scope.$root.currentUser){
    	$location.path('/');
    }
}]);