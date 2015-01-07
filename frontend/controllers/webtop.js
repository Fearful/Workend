'use strict';

angular.module('workend').controller('webtop', ['$scope', '$mdSidenav', '$mdBottomSheet', '$timeout', '$window', 'WEsession', function($scope, $mdSidenav, $mdBottomSheet, $timeout, $window, WEsession){
	$scope.toggleSidebar = function(){
		$mdSidenav('right').toggle();
	};
	$window.document.oncontextmenu = function(e){
		e.stopPropagation();
		e.preventDefault();
	    $mdBottomSheet.show({
	      templateUrl: '/api/v1/partials/optionsBottom',
	      controller: 'WEshortcuts',
	      targetEvent: e
	    });
	};
	$scope.logout = function(){
		WEsession.logout();
	};
}]);