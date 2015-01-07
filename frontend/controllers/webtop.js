'use strict';

angular.module('workend').controller('webtop', ['$scope', '$mdSidenav', '$mdBottomSheet', '$timeout', '$window', 'WEsession', '$mdDialog', function($scope, $mdSidenav, $mdBottomSheet, $timeout, $window, WEsession, $mdDialog){
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
	$scope.projects = function(e){
		$mdDialog.show({
	      controller: 'projectsDialog',
	      templateUrl: '/api/v1/partials/projectsDialog',
	      targetEvent: e,
	    });
	};
	$timeout(function(){
		if($scope.$root.currentUser){
			$mdSidenav('right').open();
		}
	},1000);
}]);