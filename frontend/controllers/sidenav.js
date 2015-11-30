'use strict';

angular.module('workend').controller('sidenav', ['$scope', '$mdSidenav', '$mdBottomSheet', '$timeout', '$window', 'WEsession', '$mdDialog', '$http', '$location', '$mdToast', '$route', '$translate', function($scope, $mdSidenav, $mdBottomSheet, $timeout, $window, WEsession, $mdDialog, $http, $location, $mdToast, $route, $translate){
	$scope.menu = [{
		name: 'Home',
		action: function(){
			$location.path(this.path);
		},
		icon: 'home',
		path: '/',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Boards',
		action: function(){
			$location.path(this.path);
		},
		icon: 'dashboard',
		path: '/boards',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Calendar',
		action: function(){
			$location.path(this.path);
		},
		icon: 'event',
		path: '/calendar',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Pr0t0',
		action: function(){
			$location.path(this.path);
		},
		icon: 'create',
		path: '/pr0t0',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Documents',
		action: function(){
			$location.path(this.path);
		},
		icon: 'bookmark',
		path: '/document',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Timesheet',
		action: function(){
			$location.path(this.path);
		},
		icon: 'schedule',
		path: '/timesheet',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Tests',
		action: function(){
			$location.path(this.path);
		},
		icon: 'thumbs_up_down',
		path: '/test',
		checkState: function(){
			return $location.path() == this.path;
		}
	}, {
		name: 'Reports',
		action: function(){
			$location.path(this.path);
		},
		icon: 'insert_chart',
		path: '/reports',
		checkState: function(){
			return $location.path() == this.path;
		}
	}];
}]);