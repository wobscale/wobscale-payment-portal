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
    // Now to load stuff; we need this user's subscriptions and related information
    var accessToken = window.localStorage.getItem("githubAccessKey");

    $scope.newUser = false;
    $scope.plans = [];

    $scope.totalSub = function() {
      return $scope.plans.reduce(function(lhs, rhs) { return lhs.Cost*lhs.Num + rhs.Cost*rhs.Num; }, 0) / 100;
    };

    $http.post("http://paypi.wobscale.website/user", {GithubAccessToken: accessToken})
      .then(function(resp) {
        $scope.githubUsername = resp.data.GithubUsername;
        if(resp.data.NewUser) {
          $scope.newUser = true;
        } else {
          $scope.stripeId = resp.data.StripeCustomerID;
          // TODO plans payment source
        }

      }, function(err) {
        if(err.data.GithubAuthError) {
          window.alert("Bad github login");
          $scope.logout();
          return;
        }
        window.alert(JSON.stringify(err));
      });

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
  }]);
})();
