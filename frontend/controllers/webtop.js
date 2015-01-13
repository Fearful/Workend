'use strict';

angular.module('workend').controller('webtop', ['$scope', '$mdSidenav', '$mdBottomSheet', '$timeout', '$window', 'WEsession', '$mdDialog', '$http', '$location', '$mdToast', function($scope, $mdSidenav, $mdBottomSheet, $timeout, $window, WEsession, $mdDialog, $http, $location, $mdToast){
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
	      targetEvent: e
	    });
	};
	function logout(){
		if($mdSidenav('left').isOpen()){
			$mdSidenav('left').toggle();
		}
		WEsession.logout();
	};
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
	$scope.chartOptions = {
      // Sets the chart to be responsive
      responsive: true,
      //Boolean - Whether we should show a stroke on each segment
      segmentShowStroke : true,
      //String - The colour of each segment stroke
      segmentStrokeColor : '#fff',
      //Number - The width of each segment stroke
      segmentStrokeWidth : 2,
      //Number - The percentage of the chart that we cut out of the middle
      percentageInnerCutout : 0, // This is 0 for Pie charts
      //Number - Amount of animation steps
      animationSteps : 100,
      //String - Animation easing effect
      animationEasing : 'easeOutBounce',
      //Boolean - Whether we animate the rotation of the Doughnut
      animateRotate : true,
      //Boolean - Whether we animate scaling the Doughnut from the centre
      animateScale : false,
      //String - A legend template
      legendTemplate : '<ul class="tc-chart-js-legend"><% for (var i=0; i<segments.length; i++){%><li><span style="background-color:<%=segments[i].fillColor%>"></span><%if(segments[i].label){%><%=segments[i].label%><%}%></li><%}%></ul>'
	};
	$scope.fileData = []
	$scope.countData = [];
	if($scope.$root.currentUser){
		$scope.$root.$watch('currentUser.starred', function(newVal, oldVal){
			if(typeof newVal === 'string' && newVal.length > 0){
				getCodeDistribution(newVal);
			}
		});
	}
	var colors = [{
		color: '#2196F3',
		highlight: '#64B5F6'
	},{
		color: '#673AB7',
		highlight: '#9575CD'
	},{
		color: '#009688',
		highlight: '#4DB6AC'
	},{
		color: '#FF9800',
		highlight: '#FFB74D'
	},{
		color: '#795548',
		highlight: '#A1887F'
	},{
		color: '#8BC34A',
		highlight: '#AED581'
	}]
	function getCodeDistribution(path){
		var index = 0;
		$http.post('/api/v1/statistics', { path: path })
		.success(function (response, status, headers, config) {
			if(response.package){
				$scope.package = JSON.parse(response.package)
			} else {
				$scope.package = false
			}
			if(response.tasks){
				$scope.$root.tasks = response.tasks;
			} else {
				$scope.$root.tasks = false;
			}
			$scope.fileData = []
			$scope.countData = [];
			angular.forEach(response.fileCount, function(file){
				$scope.fileData.push({
					value: file.count,
					color: colors[index].color,
					highlight: colors[index].highlight,
					label: file.lang
				});
				if(colors[index + 1]){
					index = index + 1;
				} else {
					index = 0;
				}
			});
			angular.forEach(response.lineCount, function(file){
				$scope.countData.push({
					value: file.count,
					color: colors[index].color,
					highlight: colors[index].highlight,
					label: file.lang
				});
				if(colors[index + 1]){
					index = index + 1;
				} else {
					index = 0;
				}
			});
			$timeout(function(){
				$scope.$apply();
			})
		})
		.error(function(error, status, headers, config) {
		 //  	$location.path('/login');
			// $mdToast.show($mdToast.simple().content('Please login or sign up'));
		});
	};
	$scope.$root.runTask = function(index){
		var task = $scope.$root.tasks[index];
		$http.post('/api/v1/tasks', { task: task.name, framework: task.icon, pathToProject: $scope.$root.currentUser.starred })
		.success(function (response, status, headers, config) {
			debugger;
		})
		.error(function(error, status, headers, config) {
		 //  	$location.path('/login');
			// $mdToast.show($mdToast.simple().content('Please login or sign up'));
		});
		$mdToast.show({
		  controller: 'taskRunToaster',
		  templateUrl: '/api/v1/partials/taskRunToaster',
		  hideDelay: 6000,
		  position: 'top right'
		});
	}
}]);