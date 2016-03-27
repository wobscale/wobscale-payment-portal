(function(){
  'use strict';
  var app = angular.module('WobscalePayments', ['ngRoute']);

  app.config(['$locationProvider', '$routeProvider', function($locationProvider, $routeProvider) {
     $locationProvider.html5Mode({
       enabled: true,
     });
    $routeProvider
      .when('/subscriptions', {templateUrl: 'partials/subscriptions.html', controller: 'SubscriptionsCtrl'})
      .when('/login', {templateUrl: 'partials/login.html', controller: 'LoginCtrl'})
      .otherwise({redirectTo: '/login'});
  }]);

  app.controller('LoginCtrl', ['$scope', '$location', '$http', function($scope, $location, $http) {
    // config
    $scope.githubClientId = "a978cb85f32000ce8f8e";

    // Check if logged in; being logged in means that you have the 'githubAccessKey' element
    var auth = window.localStorage.getItem("githubAccessKey");
    if(auth !== null) {
      // logged in, presumably. If it's invalid, the subscriptions page will bounce it
      $location.path('/subscriptions');
      return;
    }

    // Check if this is the second part of a login flow
    $scope.githubCode = $location.search().code;
    if($scope.githubCode) {
      // Second part of the login workflow
      $scope.randomCode = $location.search().state;
      $scope.step = 2;
    } else {
      var randArr = new Uint32Array(5);
      window.crypto.getRandomValues(randArr);
      var randCode = "";
      for(var i=0; i < 5; i++) {
        randCode += '' + randArr[i];
      }
      $scope.randomCode = randCode;
      $scope.step = 1;
    }


    $scope.login = function() {
      $http.post("http://paypi.wobscale.website/githubLogin", {GithubCode: $scope.githubCode})
        .then(function(resp) {
          var accessToken = resp.data.AccessToken;
          window.localStorage.setItem("githubAccessKey", accessToken);
          $location.search('code', null).path('/subscriptions');
        }, function(err) {
          $location.search('code', null);
          alert("Error logging in: " + JSON.stringify(err));
        });
    };
  }]);


  app.controller('SubscriptionsCtrl', ['$scope', '$location', '$http', function($scope, $location, $http) {
    var accessToken;

    var initialize = function() {
      accessToken = window.localStorage.getItem("githubAccessKey");

      //$scope.data.loading = true;
      $scope.data = {};
      $scope.newUser = false;
      $scope.data.plans = [];
      $scope.data.subbedPlans = [];
      $scope.data.addPlanNum = 1;
      $scope.data.addPlanName = "loading...";
      $scope.data.paymentSource = "";

      updateUserInfo();
      showAvailablePlans();
      // TODO, promises for both of those, and handle loading out here
      //.then($scope.data.loading = false)
    };

    $scope.totalSub = function() {
      return $scope.data.subbedPlans.map(function(el) { return el.Cost * el.Num; }).reduce(function(lhs, rhs) { return lhs + rhs; }, 0);
    };

    $scope.addSub = function() {
      if($scope.data.addPlanNum < 1 || $scope.data.addPlanNum > 100) {
        window.alert("Invalid plan configuration; pick a small numeric number of subscriptions");
        return;
      }
      $http.post("http://paypi.wobscale.website/addSubscription", {
        GithubAccessToken: accessToken,
        PlanName: $scope.data.addPlanName,
        PlanNum: $scope.data.addPlanNum,
      }).then(function(resp) {
        initialize();
      }, function(err) {
        alert("Plan adding failed: " + JSON.stringify(err));
      });
    };

    var showAvailablePlans = function() {
      $http.get("http://paypi.wobscale.website/plans").then(function(resp) {
        $scope.data.plans = resp.data;
        $scope.data.addPlanName = $scope.data.plans[0].Name;
      }, function(err) {
        window.alert("Unable to get plan information!");
      });
    };

    var updateUserInfo = function() {
      $http.post("http://paypi.wobscale.website/user", {GithubAccessToken: accessToken})
        .then(function(resp) {
          $scope.githubUsername = resp.data.GithubUsername;
          if(resp.data.NewUser) {
            $scope.newUser = true;
          } else {
            $scope.stripeId = resp.data.StripeCustomerID;
            $scope.data.subbedPlans = resp.data.Plans;
            $scope.data.paymentSource = resp.data.PaymentSource;
          }
        }, function(err) {
          if(err.data.GithubAuthError) {
            window.alert("Bad github login");
            $scope.logout();
            return;
          }
          window.alert(JSON.stringify(err));
        });
    };

    var createCustomer = function(token) {
      $http.post("http://paypi.wobscale.website/new", {
        Email: $scope.email,
        Nickname: $scope.nickname,
        GithubAccessToken: accessToken,
        StripeToken: token
      }).then(function(resp) {
        alert(JSON.stringify(resp));
      }, function(err) {
        alert(JSON.stringify(err));
      });
    };

    $scope.createPayment = function() {
      window.Stripe.card.createToken(document.querySelector("#payment-form"), function(status, resp) {
        alert(status);
        var customerToken = resp.id;
        createCustomer(customerToken);
      });
    };

    $scope.logout = function() {
      window.localStorage.removeItem("githubAccessKey");
      $location.path('/login');
    };

    initialize();
  }]);
})();
