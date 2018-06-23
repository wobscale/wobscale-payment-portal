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
    // TODO $window here and below
    $scope.githubClientId = window.githubClientId;

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
      // TODO $window
      $http.post(window.apiUrl + "/githubLogin", {GithubCode: $scope.githubCode})
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


  app.controller('SubscriptionsCtrl', ['$scope', '$location', '$http', '$q', function($scope, $location, $http, $q) {
    var accessToken;

    var initialize = function() {
      accessToken = window.localStorage.getItem("githubAccessKey");

      $scope.data = {};
      $scope.newUser = false;
      $scope.data.plans = [];
      $scope.data.subbedPlans = [];
      $scope.data.addPlanNum = 1;
      $scope.data.loading = true;
      $scope.data.addPlanName = "loading...";
      $scope.data.paymentSource = "";

      $q.all([updateUserInfo(), showAvailablePlans()])
      .then(function() {
        $scope.data.creatingPayment = false;
        $scope.data.addingSub = false;
        $scope.data.loading = false;
      }, function(err) {
        console.log("error: ", err);
        $scope.logout();
      });
    };

    var reinitialize = function() {
      $q.all([updateUserInfo(), showAvailablePlans()])
      .then(function() {
        $scope.data.creatingPayment = false;
        $scope.data.addingSub = false;
      }, function(err) {
        window.alert("Error loading page data: ", err);
      });
    };

    $scope.totalSub = function() {
      return $scope.data.subbedPlans.map(function(el) { return el.Cost * el.Num; }).reduce(function(lhs, rhs) { return lhs + rhs; }, 0);
    };

    $scope.addSub = function() {
      if($scope.data.addingSub) {
        return;
      }
      $scope.data.addingSub = true;
      $scope.data.idempotencyToken = Math.floor(Math.random() * 1e6).toString();
      if($scope.data.addPlanNum < 1 || $scope.data.addPlanNum > 100) {
        window.alert("Invalid plan configuration; pick a small numeric number of subscriptions");
        $scope.data.addingSub = false;
        return;
      }
      // TODO $window here and above
      $http.post(window.apiUrl + "/addSubscription", {
        GithubAccessToken: accessToken,
        PlanName: $scope.data.addPlanName,
        PlanNum: $scope.data.addPlanNum,
        IdempotencyToken: $scope.data.idempotencyToken,
      }).then(function(resp) {
        reinitialize();
      }, function(err) {
        alert("Plan adding failed: " + JSON.stringify(err));
        $scope.data.addingSub = false;
      });
    };

    var showAvailablePlans = function() {
      // TODO $window here
      return $http.get(window.apiUrl + "/plans").then(function(resp) {
        $scope.data.plans = resp.data;
        $scope.data.addPlanName = $scope.data.plans[0].ID;
      }, function(err) {
        window.alert("Unable to get plan information!");
      });
    };

    var updateUserInfo = function() {
      // TODO $window
      return $http.post(window.apiUrl + "/user", {GithubAccessToken: accessToken})
        .then(function(resp) {
          $scope.githubUsername = resp.data.GithubUsername;
          if(resp.data.NewUser) {
            $scope.newUser = true;
          } else {
            $scope.stripeId = resp.data.StripeCustomerID;
            $scope.data.subbedPlans = resp.data.Plans;
            $scope.data.paymentSource = resp.data.PaymentSource;
          }
        });
    };

    var createCustomer = function(token) {
      $http.post(window.apiUrl + "/new", {
        Email: $scope.email,
        Nickname: $scope.nickname,
        GithubAccessToken: accessToken,
        StripeToken: token
      }).then(function(resp) {
        reinitialize();
      }, function(err) {
        alert(JSON.stringify(err));
        $scope.data.creatingPayment = false;
      });
    };

    var addPayment = function(token) {
      $http.post(window.apiUrl + "/updatePayment", {
        GithubAccessToken: accessToken,
        StripeToken: token
      }).then(function(resp) {
        $scope.editPaymentSource = false;
        reinitialize();
      }, function(err) {
        alert(JSON.stringify(err));
      });
    };

    $scope.addDefaultPayment = function() {
      window.Stripe.card.createToken(document.querySelector("#add-payment-form"), function(status, resp) {
        if(resp.error) {
          alert("Stripe error: " + resp.error.message);
          return;
        }
        var customerToken = resp.id;
        addPayment(customerToken);
      });
    };

    $scope.createPayment = function() {
      if($scope.data.creatingPayment) {
        return;
      }
      $scope.data.creatingPayment = true;
      window.Stripe.card.createToken(document.querySelector("#payment-form"), function(status, resp) {
        if(resp.error) {
          alert("Stripe error: " + resp.error.message);
          $scope.data.creatingPayment = false;
          return;
        }
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
