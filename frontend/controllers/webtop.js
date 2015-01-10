'use strict';

angular.module('workend').controller('webtop', ['$scope', '$mdSidenav', '$mdBottomSheet', '$timeout', '$window', 'WEsession', '$mdDialog', '$http', function($scope, $mdSidenav, $mdBottomSheet, $timeout, $window, WEsession, $mdDialog, $http){
	$scope.toggleSidebar = function(){
		$mdSidenav('left').toggle();
	};
	$window.document.oncontextmenu = function(e){
		e.stopPropagation();
		e.preventDefault();
		if($mdSidenav('left').isOpen()){
			$mdDialog.hide();
			$mdSidenav('left').toggle();
		}
	    $mdBottomSheet.show({
	      templateUrl: '/api/v1/partials/optionsBottom',
	      controller: 'WEshortcuts',
	      targetEvent: e
	    });
	};
	function logout(){
		if($mdSidenav('left').isOpen()){
			$mdSidenav('left').toggle();
		}
		WEsession.logout();
	};
	// $http.get('/api/v1/navigate/').success(function(data){
	// 	debugger;
	// });
	function projects(e){
		$mdDialog.show({
	      controller: 'projectsDialog',
	      templateUrl: '/api/v1/partials/projectsDialog',
	      targetEvent: e
	    });
	};
	function switchProjects(e){
		$mdDialog.show({
	      controller: 'projectsListDialog',
	      templateUrl: '/api/v1/partials/projectsListDialog',
	      targetEvent: e
	    });
	};
	$scope.options = [{
		name: 'Add project',
		action: projects
	},{
		name: 'Switch Active Project',
		action: switchProjects
	},{
		name: 'Logout',
		action: logout
	}];
}]);